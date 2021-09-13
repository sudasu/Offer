# mac指令

## 环境变量配置

### /etc/profile全局环境变量失效

macOS Cataalina(10.15)后默认终端从bash变成了zsh，可输入`cho $SHELL`查看。执行完`source /etc/profile`指令后新建终端环境变量失效的解决方案：

* 执行`vim ~/.zshrc`指令
* 在.zshrc文件最后添加:`source /etc/profile`
* 最后执行`source ~/.zshrc`

### 配置环境变量

* 输入`vim ~/.zshrc`进入.zshrc文件
* 添加如export GOPATH = "/Users/lbj/go"(保留多项如export PATH = $PATH:$GOPATH:..)
* 最后`source ~/.zshrc`立即生效

### 当前窗口环境变量

`.bash_profile` 中修改环境变量只对当前窗口有效，而且需要`source ~/.bash_profile`才能使用，`.zshrc`则相当于开机启动的环境变量。

### 全局变量需要注意的

1. 修改全局变量，需要修改/etc/profile文件
2. 修改profile文件需要sudo权限
3. 由于文件只读，需要q!或wq!强制退出

## shell alias别名设置

some more ls aliases

```
alias ll='ls -alF'
alias la='ls -A'
alias l='ls -CF'
```

## 修改hosts文件

`vi /etc/hosts`

## 文件权限

chmod [参数] [ugoa(文件所有者，所有者所在组，其他用户，所有用户)][+-=(添加，删除，覆盖)][rwx(读，写，执行)]
常用参数：  
-v   //显示权限变更的详细资料  
-R  //对当前目录及子目录进行相同的权限操作

## 删除

删除指令：`rm -rf /etc/...`，一般删除文件-f，-r表示向下递归
## 查找
find path [-name|-type] filename			//注意如果想使用匹配记得加""
## ls
ls [options] [filename...]

```
options说明：
-a 				//显示所有文件及目录(.开头的隐藏文件也会列出)
-l				//显示文件类型，全新，拥有者，文件大小等详细信息
-r				//反转排列顺序
-t				//按时间顺序排了
-F				//在文件名后面加一符号，可执行文件加"*"，目录加"/"，啥也不加的应该就是普通文件
-R				//递归列出文件

-l字段的详细解析：
drwxr-xr-x    5 root    wheel      160  4 24 00:48 support-files
第一列共10位，第一位表示文档类型,d表示目录，-表示文件，l表示链接文件。后9位表示三种身份拥有的权限。
第二列表示链接书，表示有多少个文件链接到inode号码上。
第三列表示拥有者
第四列表示所属组
第五列表示文档容量大小，单位字节
第六列表示文档最后修改时间
第六列表示文档名称
```

## tail
tail [options] [file]

```
options说明：
-f 				//循环读取
-n				//显示文件尾部n行内容，不使用的话默认是10
--pid=PID		//与-f合用，表示在进程死掉后结束
s=S				//与-f合用，表示在每次读取后休眠S秒
```

## df
df [options] [file]			//查看文件系统磁盘使用情况统计

```
//options说明
-a				//包含所有具有0Blocks的文件系统
-t				//根据文件类型显示
-i				//不列出已使用block
-h				//使用易读的格式，注意计算式，1K=1000，而不是1K=1024
```

## grep
grep [match] [options]		//搜索所在行


```
-A n					//包含目标行及后n行
-B n					//包含目标行及前n行
-C n					//包含目标行及前后n行
-v						//排除目标行
--color					//高亮目标
-n						//显示第几行
-c						//目标有几行
-i						//不区分大小写
```

grep正则支持：世界上的正则表达式种类繁多且复杂，面对这样的状况，POSIX 将正则表达式进行了标准化，并把实现方法分为了两大类：基本正则表达式（BRE）扩展正则表达式（ERE）

两者的区别，更多的是元字符的区别。

在基本正则表达式（BRE）中，只承认“^”、“$”、“.”、“[”、“]”、“*”这些是元字符，所有其他的字符都被识别为普通字符。

而在扩展正则表达式（ERE）中，则在BRE的基础上增加了“（”、“）”、“{”、“}”、“？”和“+”、“|”等元字符。

最后要特别说明的一点，只有在用反斜杠\\进行转义的情况下，字符“（”、“）”、“{”和“}”才会在扩展正则表达式（ERE）中被当作元字符处理，而在基本正则表达式（ERE）中，任何元字符前面加上反斜杠反而会使其被当作普通字符来处理。这样的设计，有些奇葩，同学们一定要记清楚哦。
## mv
mv [options] src dest

```
-b				//当目标文件或目录存在时，执行覆盖前，为其创建一个副本。
-i				//询问式覆盖
-f				//直接覆盖
-n				//不覆盖
-u				//目标文件不存在或者比源文件旧时才执行移动
```

## wc
```
-c							//输出统计byte
-m							//统计字符，如果不支持多byte字符，则等同于c
-l							//输出统计行
-w							//输出单词，用空格分隔
```

## lsof(list openfile//使用man查看详细)

```
-itcp@host:[port]		//-i显示网络连接相关内容，该指令显示tcp的指定主机的port端口的连接，可拆开使用。

-s[p:s]					//如果单独使用则是强制显示文件大小，但如果和-i配合使用的话，可以用来筛选连接状态。

-u -t [username]			//-u 查看用户正在执行的，该指令可以查看该用户相关的pid，可以使用kill -9 `[command]`杀死

-c							//查看指定命令使用的文件和网络连接
-p							//查看指定进程使用的文件和网络连接
-t							//只返回pid

[filename]				//直接输入目录，显示与指定目录的交互。
+L[n]						//显示所有链接数小于n的文件，如果是1的话则是已删除但open的文件
-L							//显示no linked count？
```

## netstat(mac上并不好使，而且很卡)

```
-a							//打印出当前所有连接
-v							//打印出冗长信息，如pid之类的列信息
-p							//指定协议 如tcp
-t | -u						//分别对应tcp和udp。注意：mac上并不能这样用
-s							//打印出协议相关连接信息，处于安全因素考虑，tcp的相关信息需要root权限，如sudo netstat  -sp tcp
-n							//mac上默认是用名字来显示端口，host的，所以为了方便筛选使用该options转换为数字(马德，终于知道为什么这么卡了，每次都要dns解析host，端口协议之类的数字，使用-n就不卡了)
```

关于netstat半连接查询：可以查询SYN_RCVD状态的socket数量。全连接查询：或许可以使用establelished状态的连接数量。统计数量使用wc -l语句。
## ps

```
-A | e						//展示所有用户包含without controlling terminals的信息
-f							//打印uid,pid,parent pid,recent CPU usage,process start time,controling tty,elapsed CPU usage和associated command。如果其中使用了-u，则会将uid转换为用户名。
-a							//展示所有用户的进程信息，但是会跳过without controlling terminal(是指守护进程吗)的进程
-u [usernames] |-U [userIds]	//根据用户名查找，注意：如果使用多个匹配参数，会取交集 | 感觉linux和mac可能会不一样
-x｜X						//对已经匹配上的进程，-x会包含without controlling terminal的进程，-X会排除掉，如果同时出现按lastest的来。
-c							//将command列仅输出可以执行文件的名字，而不是整个命令行
-j							//打印user,pid,ppid,pgid,sess,jobc,state,tt,time和command信息
-l							//打印uid,pid,ppid,flags,cpu,pri,nice,vsz=SZ,rss,wchan,state=S,paddr=ADDR,tty,time和command=CMD的信息
```

## curl