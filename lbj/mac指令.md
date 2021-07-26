#mac指令
##环境变量配置
###/etc/profile全局环境变量失效
macOS Cataalina(10.15)后默认终端从bash变成了zsh，可输入<code>echo $SHELL</code>查看。执行完 source /etc/profile指令后新建终端环境变量失效的解决方案：

*  执行<code>vim ~/.zshrc</code>指令
*  在.zshrc文件最后添加:<code>source /etc/profile</code>
*  最后执行<code>source ~/.zshrc</code>

###配置环境变量
* 输入<code>vim ~/.zshrc</code>进入.zshrc文件
* 添加如export GOPATH = "/Users/lbj/go"(保留多项如export PATH = $PATH:$GOPATH:..)
* 最后<code>source ~/.zshrc</code>立即生效

###当前窗口环境变量
`.bash_profile` 中修改环境变量只对当前窗口有效，而且需要`source ~/.bash_profile`才能使用，`.zshrc`则相当于开机启动的环境变量。
###全局变量需要注意的
1. 修改全局变量，需要修改/etc/profile文件
2. 修改profile文件需要sudo权限
3. 由于文件只读，需要q!或wq!强制退出

##shell alias别名设置
<---some more ls aliases
```
alias ll='ls -alF'
alias la='ls -A'
alias l='ls -CF'
```
##修改hosts文件
<code>vi /etc/hosts</code>
##文件权限
chmod [参数] [ugoa(文件所有者，所有者所在组，其他用户，所有用户)][+-=(添加，删除，覆盖)][rwx(读，写，执行)]
常用参数：  
-v   //显示权限变更的详细资料  
-R  //对当前目录及子目录进行相同的权限操作
##删除
删除指令：`rm -rf /etc/...`，一般删除文件-f，-r表示向下递归
##查找
find path [-name|-type] filename			//注意如果想使用匹配记得加""
##ls
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
##tail
tail [options] [file]

```
options说明：
-f 				//循环读取
-n				//显示文件尾部n行内容，不使用的话默认是10
--pid=PID		//与-f合用，表示在进程死掉后结束
s=S				//与-f合用，表示在每次读取后休眠S秒
```
##df
df [options] [file]			//查看文件系统磁盘使用情况统计

```
//options说明
-a				//包含所有具有0Blocks的文件系统
-t				//根据文件类型显示
-i				//不列出已使用block
-h				//使用易读的格式，注意计算式，1K=1000，而不是1K=1024
```
##网络
查看端口使用:`lsof -i tcp:[port]`
