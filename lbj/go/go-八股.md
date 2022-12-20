# [八股](https://docs.kilvn.com/GoExpertProgramming/chapter03/3.1-go_schedule.html)

## defer

recover失效有条规则是，recover函数必须在defer方法直接调用

## 协程调度

g:go的协程，一般是平时执行的工作任务
m:工作线程，一般与p一一对应
p:概念上的处理器，主要表示go协程执行的系统资源，并带有调度的作用

go协程的状态:grunnable,gidle,gdead,gwaiting,gsyscall,grunning

grunnable可以从grunning,gwaiting,gsyscall转换而来，gdead则由grunnable转换而来

为保证不出现饥饿的情况，通过时间片轮转的方式将grunning转换成grunnable。且如果出现不同P分配的G不均衡，如一个P的所有任务都执行完成此时全局队列也没有存在G，就会发生窃取现象从其他P中窃取一半的G用来执行。

其中gsyscall这种状态是尊重系统调用的那一部分延迟设立的状态，和gwaiting作用差不多但gwaiting是由用户的channel，定时器等阻塞方法导致的。其中M数量一般会比P略大一点，所以存在M池子，当发生以上阻塞情况时P将会从池子中抽取出一个新的M，将能继续执行的g队列挂在此M上，只要P不空闲就能保证充分利用CPU。而当原来的M在系统调用或阻塞结束后被唤醒有两种情况，如果有空闲的P则会继续被分配P继续执行被阻塞任务,如果没有则将该G放入全局队列自身放入M池子休眠。

## 内存控制

## make的实现

## chan的底层

## 锁的底层

分布式锁的问题写入简历