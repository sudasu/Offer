package compute

import (
	"context"
	"encoding/json"
	"errors"
	"go-common/library/log"
	constant "pili/app/hera2/lib/const"
	"pili/app/hera2/lib/util"
	"time"
)

type CacheRst struct {
	Data    string `json:"data"`
	Message string `json:"message"`
	State   int    `json:"state"`
}

type CacheState struct {
	CacheId string `json:"cache_id"`
	Message string `json:"message"`
	State   int    `json:"state"`
}

type CacheData struct {
	StartTime string `json:"starttime"`
	EndTime   string `json:"endtime"`
	Data      string `json:"data"`
	State     int64  `json:"state"`
	Version   string `json:"version"`
}

type CacheConfig struct {
	SuccessExpir, FailedExpir, ComputingExpir int64
	Version, CacheId                          string
}

type computeParam struct {
	config         *CacheConfig
	computeParam   interface{}
	computeHandler ComputeHandler
	cacheId        string
}

type ComputeHandler func(context.Context, interface{}) (interface{}, error)

type CacheDao interface {
	// 批量获取hmget映射的map值
	HMGet(c context.Context, key string, fields ...string) (dest map[string]interface{}, err error)
	// 可能两个任务交替执行，一个任务已经创建新任务了，但另一个任务还以为版本过期执行了删除操作
	// 所以使用lua脚本实现，判断过期，然后删除的排他性操作
	AtomicGetInt64AndDelExpireHash(c context.Context, key string, validateField, valueField, validateValue string) (rst int64, err error)
	// 通过事务，赋值hash对象，且设置ttl
	SetHashValue(c context.Context, key string, obj interface{}, timeoutSec int) error
	// 由于不能对整个hash对象进行setnx，所以必须对某个关键field进行判断
	SetHashValueNX(c context.Context, key string, obj interface{}, nxKey string, timeoutSec int) (bool, error)
	// 简单删除key
	DelKey(c context.Context, key string) (err error)
}

// 批量查询计算缓存状态，message等详细信息，返回不存在的状态由两部分原因导致：1.本身不存在。2.版本落后，需要被覆盖，直接执行删除操作
func GetDetailCacheStates(ctx context.Context, cacheIds []string, vers string, d CacheDao) ([]*CacheState, error) {
	var (
		rs []*CacheState
	)
	for _, v := range cacheIds {
		// 版本落后，惰性删除
		cacheState, err := d.AtomicGetInt64AndDelExpireHash(ctx, v, constant.Redis_Version_Field, constant.Redis_State_Field, vers)
		if err != nil {
			log.Error("[GetIdcBwStatDetailCacheStates] get IdcBwStat cache version error: %v", err)
			return nil, err
		}
		r := &CacheState{
			State:   int(cacheState),
			CacheId: v,
			Message: constant.TaskDefaultMessage[int(cacheState)],
		}
		rs = append(rs, r)
	}
	return rs, nil
}

// 批量查询计算缓存状态，message等详细信息，返回不存在的状态由两部分原因导致：1.本身不存在。2.版本落后，需要被覆盖，直接执行删除操作
func GetCacheRst(ctx context.Context, cacheId string, vers string, d CacheDao) (*CacheRst, error) {
	rs := &CacheRst{}
	cData := &CacheData{}
	// 版本落后，惰性删除
	cacheState, err := d.AtomicGetInt64AndDelExpireHash(ctx, cacheId, constant.Redis_Version_Field, constant.Redis_State_Field, vers)
	if err != nil {
		log.Error("[Service][GetIdcBwStatCacheRst] get IdcBwStat cache version error: %v", err)
		return nil, err
	}
	switch cacheState {
	case constant.TaskNoRst:
		rs.Message = constant.TaskDefaultMessage[constant.TaskNoRst]
		rs.State = constant.TaskNoRst
		return rs, nil
	case constant.TaskComputing:
		rs.Message = constant.TaskDefaultMessage[constant.TaskComputing]
		rs.State = constant.TaskComputing
		return rs, nil
	case constant.TaskSuccess, constant.TaskFailed:
		// 计算任务进入终态，需要查询出计算任务的结果
		m, err := d.HMGet(ctx, cacheId, constant.Redis_Data_Field, constant.Redis_State_Field)
		if err != nil {
			log.Error("[Service][GetIdcBwStatCacheRst] get IdcBwStat cache data error: %v", err)
			return nil, err
		}
		err = util.RedisMap2Objcet(m, cData)
		if err != nil {
			log.Error("[Service][GetIdcBwStatCacheRst] convert redis map to cache data error: %v", err)
			return nil, err
		}
		// 失败的任务data里面存储的是失败信息
		if cacheState == constant.TaskFailed {
			rs.Message = cData.Data
			rs.State = constant.TaskFailed
		} else {
			rs.Data = cData.Data
			rs.State = constant.TaskSuccess
			rs.Message = constant.TaskDefaultMessage[constant.TaskSuccess]
		}
	}
	return rs, nil
}

// 注意传入的ctx，将会限制配置的任务的数据写入和计算，建议不要直接使用http的时间限制来限制计算
func LaunchCompute(ctx context.Context, options interface{}, handler ComputeHandler, configs *CacheConfig,
	d CacheDao) (*CacheState, error) {
	var cacheIds []string
	// 入参校验
	if err := paramValidate(options, handler, configs); err != nil {
		return nil, err
	}
	rst := &CacheState{
		CacheId: string(configs.CacheId),
	}
	cacheIds = append(cacheIds, rst.CacheId)
	// 得到一些经过处理过的缓存信息
	states, err := GetDetailCacheStates(ctx, cacheIds, configs.Version, d)
	if err != nil {
		log.Error("[Service][LaunchComputeIdcBwStat] query cache state err:%v", err)
		return nil, err
	}

	cp := &computeParam{
		config:         configs,
		computeParam:   options,
		computeHandler: handler,
		cacheId:        rst.CacheId,
	}

	switch states[0].State {
	case constant.TaskNoRst:
		rst.State = constant.TaskReady
		rst.Message = constant.TaskDefaultMessage[constant.TaskReady]
		// 此处error可能发生的场景在于，两个相同任务并行的执行创建。但由于使用了setnx，只有一个任务创建成功，另一个只能error了。
		// 此处的error，可能不能并入前端的轮询逻辑，而是直接放弃轮询，可以优化，但没必要。
		if err := asynCompute(ctx, cp, d); err != nil {
			log.Error("[Service][LaunchComputeIdcBwStat] async compute idc bw stat err:%v", err)
			return nil, err
		}
	case constant.TaskFailed:
		// 此处简化逻辑，直接删除失败任务，再重新发起任务，后续可更改
		// 本来不应该在此处直接删除，存在不一致风险
		// 严谨的逻辑为，人工分析错误原因，再人工删除
		d.DelKey(context.TODO(), rst.CacheId)
		rst.State = constant.TaskReady
		rst.Message = constant.TaskDefaultMessage[constant.TaskReady]
		if err := asynCompute(ctx, cp, d); err != nil {
			log.Error("[Service][LaunchComputeIdcBwStat] async compute idc bw stat err:%v", err)
			return nil, err
		}
	default:
		rst.State = states[0].State
		rst.Message = states[0].Message
	}

	return rst, nil
}

func DefaultGenerateCacheId(param interface{}, salt string) (rst string, err error) {
	cacheId, err := json.Marshal(param)
	if err != nil {
		log.Error("[DefaultGenerateCacheId] Marshal cacheId err:%v", err)
	}
	rst = string(cacheId) + salt
	return
}

func asynCompute(ctx context.Context, param *computeParam, d CacheDao) error {
	startTime := time.Now()
	// 开始执行计算，初始化cacheData
	cacheData := &CacheData{
		State:     constant.TaskComputing,
		Version:   param.config.Version,
		StartTime: startTime.Format("2006-01-02 15:04:05"),
	}
	// 初始化任务状态，写入缓存，并发时保证只能有一个任务能初始化成功
	if err := commonSaveBwStatCacheData(ctx, param.cacheId, cacheData, param.config, d); err != nil {
		log.Error("[asynCompute]save catch data err:%v", err)
		return err
	}

	// 异步执行计算任务，并把任务结果写入缓存
	go func() {
		// 此处过期时间需要另设置，与配置里的计算过期时间一致。但可能会出现，计算完成时间刚到，没有写入redis的时间了。由于redis处依然设置了过期时间，
		// 如果此处无法写入，查询处状态将会由计算中直接变为无结果
		cx, cancel := context.WithTimeout(ctx, time.Second*(time.Duration)(param.config.ComputingExpir))
		defer cancel()
		rstData, err := param.computeHandler(cx, param.computeParam)
		cacheData.EndTime = time.Now().Format("2006-01-02 15:04:05")
		if err == nil {
			var data []byte
			cacheData.State = constant.TaskSuccess
			if data, err = json.Marshal(rstData); err == nil {
				cacheData.Data = string(data)
			}
		}
		// err可能来自于计算错误，也可能来自上一步代码的转换错误
		if err != nil {
			log.Error("[asynCompute] compute data err:%v", err)
			cacheData.State = constant.TaskFailed
			cacheData.Data = err.Error()
		}
		// 此处如果失败，则导致计算状态状态无法更改。存在一定不一致性风险。但由于有过期时间的原因，可以不补充补偿逻辑
		if err = commonSaveBwStatCacheData(cx, param.cacheId, cacheData, param.config, d); err != nil {
			log.Error("[asynCompute]save catch data err:%v", err)
		}
	}()
	return nil
}

// 将缓存数据存入redis中，不同状态的缓存数据存入缓存的逻辑不同
// 在任务初始化时，需要保证相同任务只有一个任务能正常走流程，需要使用setnx，并初始化计算超时过期时间
// 任务更新为计算成功时，录入计算结果，且时间按成功逻辑修改存储时间
// 任务更新为计算失败时，录入失败原因，且时间按失败逻辑修改存储时间
func commonSaveBwStatCacheData(ctx context.Context, cacheId string, data *CacheData, config *CacheConfig, d CacheDao) (err error) {
	// 内存缓存数据的存储状态只包含，计算中，成功，失败三种状态
	switch data.State {
	case constant.TaskComputing:
		// 使用nx防止任务重复创建和覆盖
		ok, err := d.SetHashValueNX(ctx, cacheId, data, constant.Redis_Version_Field, int(config.ComputingExpir))
		if err != nil {
			log.Error("[CommonSaveBwStatCacheData] setnx hash key: %v value error: %v", cacheId, err)
			return err
		}
		if !ok {
			err = errors.New("repeat launch same task")
			log.Error("[CommonSaveBwStatCacheData] hmsetnx repeat set redis key: %v ,error : %v", cacheId, err)
			return err
		}
	case constant.TaskSuccess:
		err = d.SetHashValue(ctx, cacheId, data, int(config.SuccessExpir))
		if err != nil {
			log.Error("[CommonSaveBwStatCacheData] setnx hash key: %v value error: %v", cacheId, err)
			return err
		}
	case constant.TaskFailed:
		err = d.SetHashValue(ctx, cacheId, data, int(config.FailedExpir))
		if err != nil {
			log.Error("[CommonSaveBwStatCacheData] setnx hash key: %v value error: %v", cacheId, err)
			return err
		}
	}
	return nil
}

func paramValidate(param interface{}, handler ComputeHandler, config *CacheConfig) (err error) {
	if param == nil || handler == nil || config == nil || config.CacheId == "" || config.Version == "" {
		err = errors.New("empty config")
	}
	return
}
