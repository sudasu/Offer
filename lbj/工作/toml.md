# toml

## toml的格式

```toml
# 常用格式：
# 1. [[p1]]    表示声明数组类型的p1
# 1. [aa] 可以在该申明下写变量赋值，如isRight = true,bb=123等，这些变量都是属于aa
# 3. [aa.bb]表明aa类型的变量b，也可以如2所示赋值表明所属bb，一般在嵌套复杂变量时这样使用
# 4. [aa.bb]也可以表明map aa、key=bb，其中bb作为map的key类型最好正确，如字符串需要加""

# 例子

[[ExcelConfigParams]]
  ExcelName = "网宿CDN"
  [ExcelConfigParams.sheet]
    [ExcelConfigParams.sheet."4.直播账单（国内）"]
       type = "live"
       layout = 1
       bitUnit = 2
       monSheet = 0
      [ExcelConfigParams.sheet."4.直播账单（国内）"."mon"]
        domain = "all"
        country = "all"
        cdn = "网宿"
        monPos = [8,3]
      [ExcelConfigParams.sheet."4.直播账单（国内）"."min"]
        domain = "all"
        country = "all"
        cdn = "网宿"
        minPos = [10,1]
```

```go
//各厂商excel解析配置项
type ExcelConfig struct {
	TemplateName      string          `json:"templateName"`
	ExcelConfigParams []ExcelConfData `json:"ExcelConfigParams"`
}

type ExcelConfData struct {
	ExcelName string                   `json:"excelName"`  //excel名称
	Sheet     map[string]SheetConfData `json:"sheetNames"` //每个sheet内数据配置
}
```

## golang解析

在创建对应结构体时，变量名称需要与toml里的变量名称相同，变量类型名称无所谓

```go
// 中间件帕拉丁从路径解析
var cf ExcelConfig
flag.Set("conf", "/Users/liubangjian01/pj/work/pili/app/hades-master/configs")
paladin.Init()
err := paladin.Get("excel.toml").UnmarshalTOML(&cf)
if err != nil {
    t.Errorf("get config error: %v", err)
}


// 从request中获取文件流解析
fs, ok := c.Request.MultipartForm.File["record_list"]
if !ok || len(fs) == 0 {
    c.JSONMap(map[string]interface{}{
        "message": "上传的toml文件不正确",
    }, ecode.RequestErr)
    return
}
fh := fs[0]
file, err := fh.Open()
if err != nil {
    c.JSONMap(map[string]interface{}{
        "message": "无法打开toml文件",
    }, err)
    return
}
defer file.Close()
// 获取字节流
buf := bytes.NewBuffer(nil)
if _, err := io.Copy(buf, file); err != nil {
    c.JSONMap(map[string]interface{}{
        "message": "获取文件流失败",
    }, err)
    return
}
var list model.RecordList
if err = toml.Unmarshal(buf.Bytes(), &list); err != nil {
    c.JSONMap(map[string]interface{}{
        "message": "toml文件解析失败",
    }, err)
    return
}
```