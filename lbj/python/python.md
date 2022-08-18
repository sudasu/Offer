# python

## 切片

```python

//Object[start_index : end_index : step]

a[-4]                 //取单值，负数为逆序
a[:] | a[::]          //从左往右取全值
a[::-1]               //从右往左取全值，注意：如果start_index和end_index表达的顺序与step不符合，则取空值。

b = a[:] ,b = copy()  //浅拷贝，嵌套数组则拷贝引用
a[3:3] = 4            //插入元素
a[3:6] = [2,4,5]      //替换一部分元素,前闭后开

```