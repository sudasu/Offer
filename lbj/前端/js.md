# js基础

## promis

```js
// 创建
var promise = new Promise(function(resolve,reject) {
    syncProcess // 放入同步处理函数，进行异步处理，根据处理结果使用resolve或reject
    ...
    resolve(object) // 除了放入正常值外，也可以继续放入下一个promise对象
})

promise.then(onFulFilled,onRejected) 
or // 等价，如果then只有一个参数，则默认为onFulFilled函数
promise.then(onFulfilled).catch(onRejected)  // 其中onFulFilled的入参为resolve的返回值，onRejected的入参是reject的返回值

// 传入resolve传入Promise对象，执行冒泡
promise.then(...).then(...).catch(..)  // 错误会通过冒泡的方式，被最后一个catch捕获

// all方法，只要一个promise被rejected，p1就会rejected，此时第一个rejected返回的值会被catch。所有的promise变成fulfilled，p1才会变成
// fulfilled，并将所有的resolve组成一个数组传回去。
var p1 = Promise.all([p1,p2,p3]);

// race方法，只要一个实例改变状态，p2会跟随该势力状态改变，并返回该实例的值
var p2 = Promise.race([p1,p2,p3]);

// Promise.resolve(),Promise.reject()
var p = Promise.resolve('Hello'); // 将某一个方法转换为对应状态的promis，如果该方法具有then方法，则该不一定
```