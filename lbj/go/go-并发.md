# 并发

## channel

```go
ch := make(chan string) //无缓冲channel不能存储数据，只能起到传输作用，所以无缓冲channel的发送和接受操作时同时进行的。
ch := make(chan int, 3) //有缓冲的通道，有缓冲队列，len和cap函数用来检测队列属性

close(ch)  //关闭通道，可以用来结束range循环,也可以使用,v,ok := <- ch进行判断是否关闭
//channel关闭后不能再发送数据，不然会panic
//可以从已关闭的channel接受默认0值，所以需要ok判断
```
channel多路复用实现条件控制协程

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

// 多路复用实现超时队列
func push(q chan int, item int, timeoutSecs int) error {

    select {
    case q <- item:
        return nil
    case <-time.After(time.Duration(timeoutSecs) * time.Second):
        return errors.New("queue full, wait timeout")
    }
}

func get(q chan int, timeoutSecs int) (int, error) {
    var item int
    select {
    case item = <-q:
        return item, nil
    case <-time.After(time.Duration(timeoutSecs) * time.Second):
        return 0, errors.New("queue empty, wait timeout")
    }

}
```

## sync

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

cond := sync.NewCond(&sync.Mutex{})  //条件变量，chan一般用于一对一，cond用于条件判断，cond的使用需要自己传入互斥锁进行初始化。
cond.L.Lock() //Wait使用前需要加锁，为何如此设计，后续研究 -- 可参考java的阻塞队列
for !condition(){
	cond.Wait()   // 阻塞，阻塞队列+1。一般来说被唤醒不一定是条件被满足，所以建议放在循环里一直检查
                  // java里的condition实现阻塞队列其实也是差不多的逻辑，可重入锁加锁，然后condition判断
}
cond.L.UnLock

cond.Signal() //随机唤醒一个协程，阻塞队列减1
cond.Boardcast() //唤醒所有协程，注意所有的唤醒操作都不需要加锁
```

## context

context用于非阻塞控制多个协程，其中done和cancel的对应操作，是通过对一原子变量的close来实现，即cancel close(chan)使得done的channel关闭得以输出终止阻塞。例子如下：

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

func (c *cancelCtx) Done() <-chan struct{} {
	d := c.done.Load()
	if d != nil { // 双重校验，保证锁后对象不会重复操作，实现单例模式
		return d.(chan struct{})
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	d = c.done.Load()
	if d == nil {
		d = make(chan struct{})
		c.done.Store(d)
	}
	return d.(chan struct{})
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

## sync/atomic

var v atomic.Value
v.Store(x)
swapped := v.CompareAndSwap(old, new)  //cas原子比较,当old值与store的值一致时才会更新

## sync/pool

存储临时的对象，随时会被人为或者GC回收掉

## [errorGroup](https://www.cnblogs.com/failymao/p/15522374.html)

[bilibili errorgroup](https://pkg.go.dev/github.com/bilibili/kratos/pkg/sync/errgroup)

## 并发模型

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
