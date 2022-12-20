# 依赖注入

一般来说，依赖注入就是通过组合的方式，在运行时决定使用的组件方法。go由于并没有在运行时，反射注入的三方复杂实现，此处介绍google的wire运行前依赖注入

## [wire](https://github.com/google/wire/blob/main/_tutorial/README.md)

wire注入，是通过提前传入构造器相关的初始化方法即provider，来实现的。case如下:

```go
//+build wireinject
// wire.go文件

func InitializeEvent() Event {
    wire.Build(NewEvent, NewGreeter, NewMessage)
    return Event{}
}
```

通过加入`//+build wireinject`注释表明要使用如下provider进行组合构造，其中`return Event{}`表示将要返回一个对象的意思，给其赋值也会被忽略。然后输入`wire`指令，就会在同目录下生成如下注入代码。(注意：wire.go文件也是需要纳入版本控制的，方便以后更改生成)

```go
// wire_gen.go文件

func InitializeEvent() Event {
    message := NewMessage()
    greeter := NewGreeter(message)
    event := NewEvent(greeter)
    return event
}
```