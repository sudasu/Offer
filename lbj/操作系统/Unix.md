# Unix

## [systemd](http://www.ruanyifeng.com/blog/2016/03/systemd-tutorial-commands.html)

## 环境变量

```
setenv [变量名称] [变量值]         /在命令行修改当前进程的环境变量

export [-fnp][变量名称]=[变量设置值]  //当前登录生效。f表示变量名为函数名；n表示删除，但并不真正删除因为对当前登录有效；p表示查看所有变量。
```

## df

```
df [OP] [FILE]              //查看当前目录下的文件系统，磁盘使用率。如没有FILE参数，则是所有文件系统

-a                          //包含重复的，伪的，不可访问的文件系统
-h				            //使用易读的格式，注意计算式，1K=1024，其中如果使用-H是1000
-i --inode                  //列出inode信息而不是块使用情况
-B --block-size=SIZE        //按比例缩放，如-BM，或--block-size=1k等同于-k

```

## du

```
du [OP] [FILE]                     // 默认为递归的检索目录及子目录的磁盘使用率

-a                                 //包含文件的使用率，而不仅仅是子目录
-h                                 //human-readable，使用易读的格式，如1k 234M
-d,--max-depth=N                   //限制递归的层数，N=0则等价于-s反应当前目录的汇总情况
-X,--exclude-from=FILE             //FILE='*.o',则排除.o文件的内容
-x,--one-file-system               //跳过不同文件系统的目录
-s,--summarize                     //反应当前目录的汇总

```

## fdisk

```
fdisk [-l] 装置名称                //分盘指令，对磁盘与文件系统进行挂载
```