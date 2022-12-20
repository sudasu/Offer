# BM

## [ping](https://info.bilibili.co/pages/viewpage.action?pageId=48990574)

ping的健康检查，krato框架封装的db类型资源都有提供ping功能，建议所有资源都封装在ping检查handle里面。

## 路径参数

```
/:name/:value,通过c.Param.Get("name")方式取得
/aa/*action,c.Param.Get("action")获得。*表示后面的所有参数，且*只能使用一次在最后面
```

## 信息采集

/metrics 用于prometheus信息采集
/metadata 可以查看所有注册的路由信息
可以使用go工具进行性能分析:go tool pprof http://127.0.0.1:8000/debug/pprof/profile


curl 'http://127.0.0.1:8000/metadata' 查看路由信息

## get/set

context包含kv的map,用于context使用

## Bind

### [validate](https://pkg.go.dev/github.com/go-playground/validator/v10#section-readme)

usage:
```go
type VChangeResultReq struct {
	JobId string `json:"job_id" validate:"required"`
}
```