package dao

import (
	"context"
	"errors"
	"go-common/library/cache/redis"
	"go-common/library/conf/paladin.v2"
	"go-common/library/log"
	constant "pili/app/hera2/lib/const"
	"pili/app/hera2/lib/util"
	"pili/library/compute"
	"strconv"
	"strings"
)

type RedisHelper interface {
	compute.CacheDao
	// 获取string类型的批量key的相同field值，注意：需要保证结果与keys的顺序一致
	GetStringBatchHashFieldValue(c context.Context, keys []string, field string) (rst []string, err error)
	// 获取string类型的key的field值
	GetStringHashFieldValue(c context.Context, key string, field string) (rst string, err error)
	// int64版本
	GetInt64BatchHashFieldValue(c context.Context, keys []string, field string) (rst []int64, err error)
	GetInt64HashFieldValue(c context.Context, key string, field string) (rst int64, err error)
}

const (
	REDIS_ARGS_REPLACEHOLD = "${ARGS}"

	// 判断是否版本过期，如果过期删除，否则返回想要的值
	_hget_expire_del_script = `
	local v = redis.call('HGET',KEYS[1],ARGV[1])
	if v == nil then
		return nil
	elseif v ~= ARGV[3] then
		redis.call('DEL',KEYS[1])
		return nil
	else 
		return redis.call('HGET',KEYS[1],ARGV[2])
	end`

	// ${ARGS}用来替换不确定参数
	_hmset_and_expire = `
	return {redis.call('HMSET',KEYS[1]${ARGS}),redis.call('EXPIRE',KEYS[1],ARGV[1])}
	`

	// 判断hash field是否存在，保证每次只有一个hash对象被初始化
	// 此处不判断是否版本过期，因为之前的代码已经提前判断过了。或者在计算过程中，版本过期了，无非是付出重新计算的代价，意义不大。
	_hsetnx_and_hmset_expire = `
	if redis.call('HGET',KEYS[1],ARGV[1]) then
		return nil
	else 
		return {redis.call('HMSET',KEYS[1]${ARGS}),redis.call('EXPIRE',KEYS[1],ARGV[2])}
	end
	`
)

func NewRedis() (r *redis.Redis, cf func(), err error) {
	var (
		cfg redis.Config
		ct  paladin.Map
	)
	if err = paladin.Get("redis.toml").Unmarshal(&ct); err != nil {
		return
	}
	if err = ct.Get("Client").UnmarshalTOML(&cfg); err != nil {
		return
	}
	r = redis.NewRedis(&cfg)
	cf = func() { r.Close() }
	return
}

func (d *dao) PingRedis(ctx context.Context) (err error) {
	if _, err = d.redis.Do(ctx, "PING"); err != nil {
		log.Error("ping redis error: %v", err)
	}
	return
}

func (d *dao) buildArgsScript(startNum, countNum int, script string) (s string) {
	argsString := ""
	for i := 0; i < countNum; i++ {
		argsString += ",ARGV[" + strconv.Itoa(startNum) + "]"
		startNum++
	}
	return strings.ReplaceAll(script, REDIS_ARGS_REPLACEHOLD, argsString)
}

// 不考虑rtt，简单复用实现了
func (d *dao) GetStringBatchHashFieldValue(c context.Context, keys []string, field string) (rst []string, err error) {
	for _, v := range keys {
		var r string
		if r, err = d.GetStringHashFieldValue(c, v, field); err != nil {
			log.Error("HGET fields redis key: %v, field: %v, error: %v", v, field, err)
			return
		} else {
			rst = append(rst, r)
		}
	}
	return
}

func (d *dao) GetStringHashFieldValue(c context.Context, key string, field string) (rst string, err error) {
	var reply interface{}
	if reply, err = d.redis.Do(c, "HGET", key, field); err != nil {
		log.Error("HGET redis key: %v, field: %v, error: %v", key, field, err)
		return
	}
	if rst, err = redis.String(reply, nil); err != nil {
		if err == redis.ErrNil {
			return constant.Redis_NoRst_String, nil
		}
		log.Error("HGET redis key: %v, field: %v, error: %v", key, field, err)
	}
	return
}

func (d *dao) GetInt64BatchHashFieldValue(c context.Context, keys []string, field string) (rst []int64, err error) {
	for _, v := range keys {
		var r int64
		if r, err = d.GetInt64HashFieldValue(c, v, field); err != nil {
			log.Error("HGET fields redis key: %v, field: %v, error: %v", v, field, err)
			return
		} else {
			rst = append(rst, r)
		}
	}
	return
}

func (d *dao) GetInt64HashFieldValue(c context.Context, key string, field string) (rst int64, err error) {
	var reply interface{}
	if reply, err = d.redis.Do(c, "HGET", key, field); err != nil {
		log.Error("HGET redis key: %v, field: %v, error: %v", key, field, err)
		return
	}
	if rst, err = redis.Int64(reply, nil); err != nil {
		if err == redis.ErrNil {
			return constant.Redis_NoRst_Int, nil
		}
		log.Error("HGET redis key: %v, field: %v, error: %v", key, field, err)
	}
	return
}

func (d *dao) AtomicGetInt64AndDelExpireHash(c context.Context, key string, validateField, valueField, validateValue string) (rst int64, err error) {
	var conn = d.redis.Conn(c)
	defer conn.Close()
	//由于go-common redis工具的bug,不使用脚本封装 lua := redis.NewScript(3, _hget_expire_del_script)
	if rst, err = redis.Int64(conn.Do("eval", _hget_expire_del_script, 1, key, validateField, valueField, validateValue)); err != nil {
		if err == redis.ErrNil {
			return constant.Redis_NoRst_Int, nil
		}
		log.Error("lua script: %v excute err: %v ", _hget_expire_del_script, err)
	}
	return
}

func (d *dao) SetHashValue(c context.Context, key string, obj interface{}, timeoutSec int) error {
	var conn = d.redis.Conn(c)
	defer conn.Close()
	hashMap := util.SimpleObj2Map(obj)
	args := redis.Args{}
	for k, v := range hashMap {
		args = append(args, k)
		args = append(args, v)
	}
	params := redis.Args{}
	// 不使用伪事务，是因为集群不支持
	// 通过lua脚本的方式，单线程阻塞执行具备排他性，不具备完全的原子性，执行一半时redis崩溃是不会还原的
	// 但由于内存操作很快，一般不容易半路崩溃，除非脚本很长很慢，有一致性风险但很小
	// 如果不使用排他性操作，则HMSET操作完了，由于网络故障或者其他原因ttl可能没设置上，风险较高
	script := d.buildArgsScript(2, len(args), _hmset_and_expire)
	// 传入脚本
	params = append(params, script)
	// 传入key数量
	params = append(params, 1)
	// 传入key值
	params = append(params, key)
	// 传入时间 arg
	params = append(params, timeoutSec)
	// 传入参数
	params = append(params, args...)
	//由于go-common redis工具的bug,不使用脚本封装
	if _, err := redis.Values(conn.Do("eval", params...)); err != nil {
		log.Error("lua script: %v excute err: %v ", script, err)
		return err
	}
	return nil
}

// 建议使用lua脚本，这样就不会分成两个事务来处理该任务
// 有非常非常小的可能，第一个事务执行一半崩溃，导致没有赋值ttl，造成无法创建任务，第二个事务可以被覆盖所以不会产生该问题
// 对上述问题redis集群可能有补偿逻辑，暂时放置
func (d *dao) SetHashValueNX(c context.Context, key string, obj interface{}, nxKey string, timeoutSec int) (bool, error) {
	var conn = d.redis.Conn(c)
	defer conn.Close()
	hashMap := util.SimpleObj2Map(obj)
	_, ok := hashMap[nxKey]
	if !ok {
		err := errors.New("nxKey doesn't exist")
		log.Error("HMSETNX redis key: %v err: %v", key, err)
		return false, err
	}
	args := redis.Args{}
	params := redis.Args{}
	for k, v := range hashMap {
		args = append(args, k)
		args = append(args, v)
	}
	// 不使用伪事务，是因为集群不支持
	// 通过lua脚本的方式，单线程阻塞执行具备排他性，不具备完全的原子性，执行一半时redis崩溃是不会还原的
	// 但由于内存操作很快，一般不容易半路崩溃，除非脚本很长很慢，有一致性风险但很小
	// 如果不使用排他性操作，则HMSET操作完了，由于网络故障或者其他原因ttl可能没设置上，风险较高
	script := d.buildArgsScript(3, len(args), _hsetnx_and_hmset_expire)
	// 传入脚本
	params = append(params, script)
	// 传入key数量
	params = append(params, 1)
	// 传入key值
	params = append(params, key)
	// 校验存在的field值
	params = append(params, nxKey)
	// 传入时间 arg
	params = append(params, timeoutSec)
	// 传入参数
	params = append(params, args...)
	//由于go-common redis工具的bug,不使用脚本封装
	if _, err := redis.Values(conn.Do("eval", params...)); err != nil {
		if err == redis.ErrNil {
			return false, nil
		}
		log.Error("lua script: %v excute err: %v ", script, err)
		return false, err
	}
	return true, nil
}

func (d *dao) HMGet(c context.Context, key string, fields ...string) (dest map[string]interface{}, err error) {
	//检测参数
	if len(fields) == 0 {
		err = errors.New("fields can't be nil")
		log.Error("HMGet redis key: %v, err: %v", key, err)
		return
	}
	//拼接查询参数+ redis请求
	args := redis.Args{}
	args = args.Add(key)
	for _, item := range fields {
		args = args.Add(item)
	}
	conn := d.redis.Conn(c)
	defer conn.Close()
	var rst []interface{}
	if rst, err = redis.Values(conn.Do("HMGET", args...)); err != nil {
		log.Error("HMGet redis key: %v get fields, err: %v", key, err)
		return
	} else {
		dest = make(map[string]interface{})
		for i := 0; i < len(fields); i++ {
			dest[fields[i]] = rst[i]
		}
	}
	return
}

// MGet get multi key value
func (d *dao) MGet(c context.Context, keys ...string) (dest []string, err error) {
	//检测参数
	if len(keys) == 0 {
		return
	}
	//拼接查询参数+ redis请求
	args := redis.Args{}
	for _, item := range keys {
		args = args.Add(item)
	}
	conn := d.redis.Conn(c)
	defer conn.Close()
	if dest, err = redis.Strings(conn.Do("MGET", args...)); err != nil {
		log.Errorc(c, "MGet conn.Do(mget) error(%v) args(%+v)", err, args)
		return
	}
	return
}

// DelKey redis
func (d *dao) DelKey(c context.Context, key string) (err error) {
	var conn = d.redis.Conn(c)
	defer conn.Close()
	_, err = conn.Do("DEL", key)
	if err != nil {
		log.Error("del redis key: %v err: %v", key, err)
		return err
	}
	return
}
