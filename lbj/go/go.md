# [golang](https://golang.google.cn/doc/)

## [go语言规范](https://info.bilibili.co/pages/viewpage.action?pageId=176564979#%E8%A1%A8%E9%A9%B1%E5%8A%A8%E6%B5%8B%E8%AF%95)

## 初始化

GOPATH表示go项目的工作目录,go get获取的项目都放在此工作目录
GOBIN表示go编译生成的程序安装目录，go install命令会将go程序安装至GOBIN目录下供终端使用

`go mod init {modulename}`初始化模块
`go mod tidy` 增加缺失的包，移除没用的包

使用go build ./main.go 将会编译生成main的可执行文件
go install ./main.go 也可以将程序生成在GOBIN目录，供终端使用

为了跨平台使用，使用GOOS和GOARCH代表目标操作系统(linux,windows,darwin)和目标处理器架构(amd64,arm64,386)
`GOOS=darwin GOARCH=amd64 go build ./main.go`

## 变量

1. go语言中只允许布尔类型，字符串等基础类型作为常量
2. iota是一个常量生成器，用法如下
   
```go
const (
  one = 1
  two = 2
  three = 3
)

//使用 iota
const (
  one = iota+1
  two
  three
)
```

3. 字符串

```go
//strconv包常用来进行数字，字符串转换
i := 10
itos := strconv.Itoa(i)
stoi,err := strconv.Atoi(itos)

//strings常用字符串工具包
HasPrefix(s1,"H"))
//在s1中查找字符串o
Index(s1,"o"))
//把s1全部转为大写
ToUpper(s1))

//string可以直接通过[]byte转换，注意中文在UTF8编码下对应3个字节，所以计算长度时需要注意，可以使用
//utf6.RuneCountinString函数统计，这样就不会出现问题。for range时是使用unicode编码的，一个汉字只占一个字宽
```

4. 字符类型

```go
var ch byte = 65         //byte是uint8的别名
var ch byte = '\x41'     //\x表示16进制数，\表示8进制数
//go语言中字符用int来表示,被称为runes。在书写unicode字符时,使用\u前缀表示4字节,使用\U表示8字节
```

## 循环，分支

go语言为防止写break，case默认自带break，如确实需要执行下个case,加入fallthrough执行下**一个**cass
```go
switch j:=1;j{
  case 1:
	fmt.Println("2")
    fallthrough
  case 2:
    fmt.Println("1")
  default:
    fmt.Println("无匹配")
}
```

## 类型转换

golang是强类型语言，但不提供隐式类型转换(比较坑，int和int32这种都需要自己转换)，只提供强制类型转换和断言。

```go
//强制类型转换:type(v),float和int可以强转。int和byte可以强转，用作字符的使用
var a int32 = 10
var b int64 = int64(a)
//指针中的强制类型转换需要通过unsafe包中的函数实现
var a int =10
var b *int =&a
var c *int64 = (*int64)(unsafe.Pointer(b))

//类型断言
//value,ok := interface{}.(type),使用类型断言的变量必须是空接口
//如果type为具体某个类型，检查成功则返还类型为type的动态值
//如果type为接口，则不回返回动态值，只会返回接口值
//如果动态值时nil，则类型断言一定会失败
```

## 接口

1. 当值类型作为接收者，person类型和*person类型都实现了该接口。
2. 当指针类型作为接收者，只有 *person类型实现了该接口
3. 指针类型接受者，可以改变接受者的值，因为接收者参数是拷贝的副本

## [反射例子](https://cloud.tencent.com/developer/article/1864032)

## error和panic

```go
//最简单的error，使用error.New("error了")就可以生成了，也可以接口自定义实现包含更多信息
fun(t* type) Error() string
//func panic(v interface),panic如果不进行恢复将会导致程序崩溃，使用以下方法进行捕获下层的panic
defer func() {
      if p := recover(); p != nil {
         fmt.Println(p)
      }
}()
```

## 并发

### channel和sync

```go
ch := make(chan string) //无缓冲channel不能存储数据，只能起到传输作用，所以无缓冲channel的发送和接受操作时同时进行的。
ch := make(chan int, 3) //有缓冲的通道，有缓冲队列，len和cap函数用来检测队列属性

close(ch)  //关闭通道，可以用来结束range循环,也可以使用,v,ok := <- ch进行判断是否关闭
//channel关闭后不能再发送数据，不然会panic
//可以从已关闭的channel接受默认0值，所以需要ok判断
```
channel实现条件控制协程

```go
for {
	select {
		case <- stopWK:
			fmt.Println("停止工作")
			return
		default:
			fmt.Println("继续工作")
	}
}
```

```go
var mutex = sync.Mutex{}   //互斥锁
var rwmutex = sync.RWmutex{} //读写锁

var wg sync.WaitGroup //协调任务完成，但实际chan的阻塞等待应该也差不多吧，更简单点
wg.Add(1)  //添加一个计数器
go func(){
	wg.Done() //计数器减1
}
wg.Wait() //等待至计数器清零

var once sync.Once //在高并发的情况下保证只执行一次
for i:0;i<10;i++{  //是个并发只有一个任务成功执行了
	go func ()  {
		once.Do(something)
	}
}

cond := sync.NewCond(&sync.Mutex{})  //条件变量，chan一般用于一对一，cond用于1对多，cond的使用需要自己传入互斥锁进行初始化。
cond.L.Lock() //Wait使用前需要加锁，为何如此设计，后续研究
for !condition(){
	cond.Wait()   //阻塞，阻塞队列+1。一般来说被唤醒不一定时条件被满足，所以建议放在循环里一直检查
}
cond.L.UnLock

cond.Signal() //随机唤醒一个协程，阻塞队列减1
cond.Boardcast() //唤醒所有协程，注意所有的唤醒操作都不需要加锁
```

### context

context用于非阻塞控制多个协程，例子如下：

```go
import (
 "context"
 "fmt"
 "sync"
 "time"
)
func main() {
 var wg sync.WaitGroup
 ctx, stop := context.WithCancel(context.Background())  //得到控制的contex和取消函数
 wg.Add(1)
 go func() {
  defer wg.Done()
  worker(ctx)
 }()
 time.Sleep(3*time.Second) //等待3秒
 stop() //等待完成后，发出取消指令
 wg.Wait()
}

func worker(ctx context.Context){
 for {
  select {
  case <- ctx.Done(): //当contex取消后，所有的协程的该通道都会收到通知
   fmt.Println("下班咯~~~")
   return
  default:
   fmt.Println("认真摸鱼中，请勿打扰...")
  }
  time.Sleep(1*time.Second)
 }
}
```

context接口如下

```go
type Context interface {
   Deadline() (deadline time.Time, ok bool)  //返回的deadline是截止时间，到了这个时间context自动发起取消，返回值ok表示是否设置了截止时间
   Done() <-chan struct{}  //这种写法表示只能channel输出，在被读取时说明已经发出了取消信号，可以做清理退出的操作了
   Err() error //输出被取消的原因
   Value(key interface{}) interface{} //context上绑定的值
}
```

context树的根节点，通过context.Background获得。go语言也提供了emptyCtx(一般不用，不能取消，不能设置超时等功能)和TOBO(用法和Background一样，但Background一般用在主函数，TODO一般用在不确定功能时)

context树的生成,使用如下context的方法

```go
WithCancel(parent Context)  //生成一个可取消的 Context。
WithDeadline(parent Context, d time.Time)  //生成一个可定时取消的 Context，参数 d 为定时取消的具体时间。
WithTimeout(parent Context, timeout time.Duration)  //生成一个可超时取消的 Context，参数 timeout 用于设置多久后取消
WithValue(parent Context, key, val interface{})  //生成一个可携带 key-value 键值对的 Context。
```

context使用原则：

* Context 不要放在结构体中，需要以参数方式传递
* Context 作为函数参数时，要放在第一位，作为第一个参数
* 使用 context。Background 函数生成根节点的 Context
* Context 要传值必要的值，不要什么都传
* Context 是多协程安全的，可以在多个协程中使用

### 并发模型

单条件控制模型：

```go
for { //for 无限循环
  select {
    //通过 channel 控制
    case <-done:
      return
    default:
      //执行具体的任务
  }
}

for _, s := range []int{} { //for 有限循环，可以类比使用waitGroup
   select {
   case <-done:
      return
   case resultCh <- s:
   }
}

timeout := time.After(3*time.Second) //返回一个3秒后输出的chan
for { //for 超时退出循环
  select {
    //通过 channel 控制
    case v <- result:
		fmt.Println(“输出处理结果”)
    return
	case <- timeout //3秒后被输出，注意不要
	return error.New("超时啦")
    default:
      //执行具体的任务
  }
}
```

多条件控制

```go
ctx,stop := context.WithCancel(context.Background,3*time.Second)
go func ()  {
	worker(ctx,"打工人1")
}
go func ()  {
	worker(ctx,"打工人1")
}

func worker(ctx context.Context,name string)  {
	for{
		select {
			case <- ctx.Done()
			fmt.Println("下班")
			return
			default:
			fmt.Println("上班")
		}
	}
}
```

pipeline模式:通过channel的单向输出模式，协调多个任务到顺序进行

```go

func main() {
 coms := buy(10)    //采购10套零件
 phones := build(coms) //组装10部手机
 packs := pack(phones) //打包它们以便售卖
 //输出测试，看看效果
 for p := range packs {
  fmt.Println(p)
 }
}

//工序1采购
func buy(n int) <-chan string {
 out := make(chan string)
 go func() {
  defer close(out)
  for i := 1; i <= n; i++ {
   out <- fmt.Sprint("零件", i)
  }
 }()
 return out
}

//工序2组装
func build(in <-chan string) <-chan string {
 out := make(chan string)
 go func() {
  defer close(out)
  for c := range in {
   out <- "组装(" + c + ")"
  }
 }()
 return out
}

//工序3打包
func pack(in <-chan string) <-chan string {
 out := make(chan string)
 go func() {
  defer close(out)
  for c := range in {
   out <- "打包(" + c + ")"
  }
 }()
 return out
}

```

扇形模式:将多个并行执行的任务合并为一个结束通道

```go
func merge(ins ... <-chan string) <-chan string{
  var wg sync.WaitGroup
  out := make(chan string)
  do := func (c <-chan string)  {
    defer wg.Done()
    for s := range c{
      //do something
      out <- s
      ... 
    }
  }
  for _, v := range ins {
    wg.Add(1)
    go do(v)
  }
  go func ()  {
    wg.Wait()
    close(out)
  }()
  return out
} 
```

```go

s1 := "ssssssss"
s2 := "aaaaaaa"
done := make(chan string)
do = func (st string,cs1 <-chan string,cs2 <-chan string)  {
  c := char[](st)
  int i = 0;
  for {
    select{
      case <-done:
      break
      case <-cs1:
      if i<len(c)-1 {
        append(c[i])
        cs2 <- "1"
      }else{
        done <- "1"
        break
      }
    }
  }
  for i<len(c)-1 {
    append(c[i])
  }
}

do()

```

## 指针,make,new


指针类型的变量如果没有分配内存，默认是nil。

```go
var p *string
*p = "aaa"  //此处将会报错，因为没有分配内存
//正确使用方法是先通过new分配内存，然后再进行赋值
p = new(string)
*p = "aaa"
```

make函数是new函数的封装，方便使用者

1. `make(map[string]string)`
2. `make([]int, 2)`
3. `make([]int, 2, 4)`

第一种情况只能用在map或者chan的场景，返回长度空间默认为0。第二种情况用于指定数组长度，如上返回长度为2的slice。第三种用法，第二个参数指定切片长度，但第三个参数用来指定预留空间。即提前分配好长度足够的内存空间4，但先只能使用2的空间。

## [测试](https://geektutu.com/post/quick-go-test.html)

### 单元测试

目录结构一般如下，在对应文件下创建文件名_test名的测试文件:

```
[ceshi]
  |--[gotest]
          |--unit.go
  |--[gotest_test]
          |--unit_test.go
```

提供参数t *testing.T，可以通过t.Log的方式打印出对应日志，如:

|方法|释义|
|:---:|:---:|
|Log|打印日志|
|Logf|格式化打印日志|
|Error|打印错误日志|
|Errorf|格式化打印错误日志|
|Fatal|打印致命日志|
|Fatalf|格式化打印致命日志|
|FatalNow|打印日志后，并立刻结束程序|

go test 参数
-bench regexp 执行相应的 benchmarks，例如 -bench=.；
-cover 开启测试覆盖率；
-run regexp 只运行 regexp 匹配的函数，例如 -run=Array 那么就执行包含有 Array 开头的函数；
-v 显示测试的详细命令，如该路径下有多个测试函数，逐一输出测试信息。

bench测试提供b *testing.B参数，目录结构如下:

```
[ceshi]
  |--[gotest]
          |--benchmark.go
  |--[gotest_test]
          |--benchmark_test.go
```

文件名必须以“_test.go”结尾；
函数名必须以“Benchmark”开始；
使用命令“go test -bench=.”开始性能测试；
其中b.N随机提供测试次数，输出操作系统,cpu架构信息，cpu核数以及总共完成多少次测试,每次测试的操作时间

```go
//bech测试
func BenchmarkSlice2(b *testing.B) {
 for i := 0; i < b.N; i++ {
  gotest.Slice2()
 }
}
```

### Main测试

一般用来执行各个测试程序的调度，统一初始化等

```go
// TestMain 用于主动执行各种测试，可以测试前后做setup和tear-down操作
func TestMain(m *testing.M) {
 fmt.Println("TestMain setup.")
 retCode := m.Run() // 执行测试，包括单元测试、性能测试和示例测试
 fmt.Println("TestMain tear-down.")
 os.Exit(retCode)
}
```

## sort 

```golang
//例子

//go的排序需要定义类型，并实现三个接口
type ChangeRecordInfos struct {
	ChangeDate string              //变更记录日期
	RecordInfo []*ChangeRecordInfo //变更记录信息
}

type RecordListSortSlice []*ChangeRecordInfos

func (sl RecordListSortSlice) Len() int { return len(sl) }
func (sl RecordListSortSlice) Less(i, j int) bool { return sl[i].ChangeDate < sl[j].ChangeDate }
func (sl RecordListSortSlice) Swap(i, j int) { sl[i], sl[j] = sl[j], sl[i] }

//使用
sort.Sort(sort.Reverse(model.RecordListSortSlice(infos)))  //其中reverse返回的是一个呗interface包装的slice，表明需要逆转排序
```

## time

```go	
startTime := time.Now()
//使用Format按格式转换字符串，完整为2006-01-2 15:04:05,数字不变其他符号可以替换成想要的格式
startTime.Format("2006-01-02") 
//必须转换成本地时间，不然直接使用Parse()函数默认是utc时间
loc, _ := time.LoadLocation("Asia/Shanghai") 
stringDate := "2006-02-02"
//time.Now()输出的是cst时区时间，time.Parse()是utc，相差8小时。但无论
//是utc还是cst，当使用format(format时，应该会把时区参数加上再format)
//时都会得到正确的字符串，但time的unix值不同
endTime, _ := time.ParseInLocation("2006-01-02", stringDate, loc) 

//time.Duration虽然是int,在直接使用数字时可以不用转换，但经过计算后需要使用显式强制类型转换
a := start.Minute() + 40
start = start.Add(time.Duration(a) * time.Minute)
```

```
ticker := time.NewTicker(time.Second)
defer tick.Stop()
for{
  select{
    case t := <-ticker.C:
      fmt.Println("Current time: ",t)
  }
}
```

## runtime

```go
runtime.Gosched()      //Gosched:让出当前线程cpu，进入就绪状态
runtime.NumCPU()       //返回当前系统CPU核数
runtime.GOMAXPROCS()   //GOMAXPROCS设置最大的可同时使用的CPU核数。
runtime.Goexit()       //Goexit退出当前goroutine（但是defer语句会照常执行）。
runtime.NumGoroutine() //NumGoroutine返回正在执行和排队的任务总数。
os := runtime.GOOS     //GOOS目标操作系统。
```

## sync/atomic

var v atomic.Value
v.Store(x)
swapped := v.CompareAndSwap(old, new)  //cas原子比较,当old值与store的值一致时才会更新

## sync/pool

存储临时的对象，随时会被人为或者GC回收掉

## [metrics](https://zhuanlan.zhihu.com/p/390439038)