# vue

## 钩子函数

```js
// vue实例创建前的钩子函数，此时vue中的data和methods都还没被初始化
beforeCreate(){
    ...
}
// 此时vue中的data和methods已经创建好，是操作vue实例最早的钩子
created(){
    ...
}
// vue实例开始编辑模板，但模板只是被渲染好，还并未被挂载到真正的页面中。如模板中{{msg}}中的值，还未被赋值
beforeMount(){
    ...
}
// 此时模板已经被渲染好，如echarts等插件也可以开始被赋值了
mounted(){
    ...
}
beforeUpdate(){
    ...
}
updated(){
    ...
}
// vue执行销毁前的钩子函数
beforeDestroy(){
    ...
}
```