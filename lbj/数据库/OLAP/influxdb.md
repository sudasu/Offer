# influxdb

## 概述

InfluxDB 是一个开源分布式时序、事件和指标数据库。使用 Go 语言编写，无需外部依赖。其设计目标是实现__分布式__和__水平伸缩扩展__。InfluxDB 包括用于__存储__和__查询__数据(较少___修改___和___删除___)，在后台处理ETL或监视和警报目的，用户仪表板以及可视化和探索数据等的API。

![influxdb](https://www.influxdata.com/wp-content/uploads/influxdb-circle-products.png)

doc:
[paper](https://www.influxdata.com/_resources/techpapers-new/)

## 特点

* 专为时间序列数据编写的自定义高性能数据存储。__TSM引擎__允许高摄取速度和数据压缩
* 完全用 Go 语言编写。 它编译成单个二进制文件，没有外部依赖项
* 简单，高性能的写入和查询HTTP API
* 插件支持其他数据提取协议，如Graphite，collectd和OpenTSDB
* 专为类似SQL的查询语言量身定制，可轻松查询聚合数据
* tag允许对系列进行索引以实现快速有效的查询
* 保留策略有效地自动使过时数据过期
* 连续查询自动计算聚合数据，以提高频繁查询的效率

# #内部结构

### 存储引擎
####目的
数据在确保准确性和性能的条件下，安全写入磁盘，查询的数据返回完整正确。  
存储引擎包含以下组件：

* [预写日志 (WAL)](#1.1)
* [缓存](#1.2)
* [时间结构合并树 (TSM)](#1.3)
* [时间序列指数 (TSI)](#1.4)

#### 磁盘写入

数据通过api借助[线路协议](#2.1)写入数据(<u>一般常用HTTP POST请求？</u>)，此时存储引擎开始处理数据，然后再写入物理磁盘。成批的point被发送到influxDB，压缩并写入[预写日志 (WAL)](#1.1)实现<u>即时持久化</u>(<s>啥意思？指写入内存吗？</s>)。point在被写入内存缓存并立即变得可查询，内存缓存以[TSM](#1.3)文件的形式定期写入磁盘。随着TSM文件的积累，存储引擎将会把累积的文件继续合并压缩为更高级的TSM文件。

><table><tr><td bgcolor=#ddffdd>虽然积分可以单独发送，但为了提高发送效率，大多数应用都会选择批量发送积分。POST body中的point可以来自任意数量的<b>series, measurements</b>和<b>tag sets</b>，不必来自相同的measurements或tag sets</td></tr></table>

<h4 id="1.1">预写日志(WAL)</h4>
预写日志维持着influxDB数据，当存储引擎重新启动时，保证数据在发生意外故障时是持久的。当存储引擎收到写请求时会发生以下步骤：

1. 写入请求附加到 WAL 文件的末尾。
2. 使用[<code>fsync()</code>](#120.1)将数据写入磁盘。
3. 内存缓存已更新。
4. 当数据成功写入磁盘时，响应确认写入请求成功。

<code>fsync()</code>获取file并且在整个过程中通过pending writes写入磁盘。作为系统调用，<code>fsync()</code>有着高昂的内核上下文切换的开销，但是保证了数据在磁盘上的安全性。当存储引擎重新启动时，WAL文件被读回内存数据库，InfluxDB然后响应对<code>/read</code>端点的请求

[...](#120.2)

<h4 id="1.2">缓存</h4>
The cache是存储在WAL中当前内存中point数据的拷贝。缓存将会：
1. 根据key(measurement, tag set, and unique field)组织point存储在自己的时间顺序范围中。
2. 存储未压缩的数据。
3. 每次存储引擎重启时都会从WAL中获取更新。缓存在runtime时查询和合并存储在TSM文件中的数据。

<h4 id="1.3">TSM</h4>
为了高效的压缩和存储数据，存储引擎通过series key对field数据进行分组，然后通过时间将其排序。  
存储引擎采用TSM(时间结构合并树)的数据结构。TSM文件以columnar的格式压缩series数据，为了提高效率，存储引擎只存储series之间的差值(或者说deltas)。面向列的存储让引擎通过series key读取而忽略了无关的数据。  
在filed被安全的存储在TSM文件中后，WAL将被truncate和缓存将会被清除。

<h4 id="1.4">时间序列索引</h4>
当数据的基数(series的数量)开始增长，将查询读取更多的series keys同时也会变得更慢。时间序列索引存储着series keys，确保当数据基数增长时查询仍然很快。
###文件系统布局
####influxDB文件结构
引擎路径：当influxDB存储时序数据存储引擎使用的如下目录
>* data:存储时间结构合并树(TSM)文件
>* wal:存储写前日志(WAL)文件

为了自定义此路径，请使用engine-path选项去配置

Bolt路径：Boltdb数据库的文件路径，非时间序列数据的基于文件的键值存储，例如 InfluxDB 用户、仪表板、任务等。要自定义此路径，请使用bolt-path 配置选项。

Configs路径：influx CLI连接配置(configs)的文件路径。要自定义此路径，请在influx CLI命令中使用该<code>--configs-path</code>标志。

InfluxDB 配置文件：一些操作系统和包管理器在磁盘上存储默认的InfluxDB(<code>influxd</code>) 配置文件。有关使用InfluxDB配置文件的更多信息，请参阅[配置选项](#1.6)。

#### 文件布局

>~/.influxdbv2/
>>engine/
>>>data/TSM directories and files  
wal/WAL directories and files

>>configs(注意是influx CLI连接配置)  
influxd.bolt

<h3 id="1.6">配置选项</h3>
自定义influxDB的配置可以通过使用<u><code>influxd</code>配置命令</u>，<u>设置环境变量</u>或者在<u>配置文件</u>汇总定义配置选项来实现。

<h4 id="1.6.3">influxDB配置文件</h4>
当influxd启动是，它会检查在当前工作目录中的config.*文件。config文件支持多种语法配置，如YAML(.yaml,.yml)，TOML(.toml)，JSON(.json)。为了自定义配置文件的目录路径，可以通过设置INFLUXD_CONFIG_PATH环境变量来定义。

<code>export INFLUXD_CONFIG=/usr/local/etc/influxdv2</code>

在<code>influxd</code>启动时，会先检查INFLUXD_CONFIG_PATH目录。
####配置方式举例([详细配置](https://docs.influxdata.com/influxdb/v2.0/reference/config-options/#configuration-options))
配置用于存储存储引擎文件位置的engine-path：
>
**Default**:<code>~/.influxdbv2/engine</code><br/>
**influxd flag**:<code>influxd --engine-path=~/.influxdbv2/engine</code><br/>
**environment variable**:<code>export INFLUXD_ENGINE_PATH=~/.influxdbv2/engine</code><br/>
**configuration file**:<code>engine-path: /users/user/.influxdbv2/engine</code><br/>



<details>
<summary>展开查看</summary>
<pre><code>
System.out.println("Hello to see U!");
</code></pre>
</details>

<div id="2.1">线路协议</div>
### 相关资料
<div id="120.1"><h3><code>fsync()</code>系统调用相关</h3></div>

#### 磁盘同步函数
<code>sync()</code>、<code>fsync()</code>、<code>fdatasync()</code>都是linux提供的磁盘同步函数，分别有以下特点。  
>
* <code>sync()</code>函数是将所有修改过的块缓冲区排入写队列，然后就返回，它并不等待实际写磁盘操作结束，通常被称为update的系统守护进程周期性(一般每隔30秒)调用。这就保证了定期冲洗内核的块缓冲区。命令sync(1)也调用sync函数。  
* <code>fsync()</code>函数只对由文件描述符fd指定的单一文件起作用，并且pending write到磁盘操作结束，然后返回。fsync可用于数据库这样的应用程序，需要确保修改过的块立即写到磁盘上。  
* `fdatasync()`类似于`fsync()`，但只影响文件的数据部分，而fsync除数据外还会同步更新文件的属性。

#### 为什么需要磁盘同步函数

传统的UNIX实现在内核中设有__缓冲区高速缓存__或__页面高速缓存__，大多数磁盘I/O都通过缓冲进行。当将数据写入文件时，内核通常先将该数据复制到其中一个缓冲区中，如果该缓冲区尚未写满，则并不将其排入<font color="#ff0000">__输出队列__</font>，而是等待其写满或者当内核需要重用该缓冲区以便存放其他磁盘块数据时，再将该缓冲排入（调用<font color="#ff0000"><code>fflush()</code></font>命令）输出队列，然后待其到达队首时，才进行实际的I/O操作。这种输出方式被称为__延迟写（delayed write）__。

延迟写<u>减少了磁盘读写次数</u>，但是却<u>降低了文件内容的更新速度</u>，使得欲写到文件中的数据在一段时间内并没有写到磁盘上。当系统发生故障时，这种<u>延迟可能造成文件更新内容的丢失</u>。为了保证磁盘上实际文件系统与缓冲区高速缓存中内容的一致性，UNIX系统提供了sync、fsync和fdatasync三个函数。
####<code>write()</code>和<code>fsync()</code>区别
[continue....](https://blog.csdn.net/hmxz2nn/article/details/82868980)

[and....](https://www.jb51.net/article/101062.htm)

[directIO](https://blog.csdn.net/alex_xfboy/article/details/91865675)(以后再写--->抄😂)

<div id="120.2">[WAL细节比较](https://zhuanlan.zhihu.com/p/137512843)