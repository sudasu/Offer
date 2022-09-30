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

## sftp

```
//连接
sftp -P 22222 <username>@relay.bilibili.co    //不加-P默认是22端口

-B  //buffer_size，制定传输 buffer 的大小，更大的 buffer 会消耗更多的内存，默认为 32768 bytes；
-P  //port，制定连接的端口号；
-R  //num_requests，制定一次连接的请求数，可以略微提升传输速度，但是会增加内存的使用量。
//连接后
get filename [newname]   //获取文件,参数:-afPr.-r可以递归获取目录下的所有文件(注意不会跟随符号连接去获取)，-a恢复之前断开的传输，注意已传输部分如果更改可能会引起文件损坏，
                         //-f将会使用fsync函pending write的方式刷盘至磁盘,-p完整复制文件的权限和访问时间
put                      //用法与get一致
lpwd                     //可以看本地目录，以免自己忘了
```

## [kill信号](https://zhuanlan.zhihu.com/p/113876980)

SIGKILL       kill -9    Term    Kill signal,不能捕获
SIGTERM            15    Term    Termination signal,可以被忽略的kill信息,kill的默认缺省值
SIGSTOP            19    Term    Stop the process,不能捕获
SIGCONT            18    Term    continue the process
SIGTSTP     ctrl+z/20    Term    暂停进程，可以被进程忽略
SIGINT       ctrl+c/2    Term    程序终止(interrupt)信号终止进程，可以被进程忽略
SIGQUIT      ctrl+\/3    Term    SIGQUIT退出时会产生core文件, 在这个意义上类似于一个程序错误信号
SIGHUP              1    Term    用户终端连接结束时发出，可以被忽略，如wget在用户退出登陆后也能继续执行。与终端脱离关系的
                                 守护进程，这个信号用鱼通知它重新读取配置文件

EOF          ctrl+c EOF  Term    (发送特殊符号EOF,通常会结束该进程，如mysql客户端连接)[https://www.cnblogs.com/jiangzhaowei/p/8971265.html]
