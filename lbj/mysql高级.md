# [mysql](https://dev.mysql.com/doc/refman/5.7/en/innodb-deadlocks-handling.html)

## 基础

mysql的模式匹配:`like`语句支持%以匹配一个或多个,`_`匹配一个，也可以通过`regexp`来使用正则表达式。为使得`regexp`区分大小写可以使用`binary`转化成二进制字符串，如`regexp binary '^b'`。(^:以...开头，$:以...结尾,.:任意一个字符,{n}:前一个规则重复5次。)  
mysqld\_safe:在UNIX上启动mysqld服务器的推荐方法，mysqld\_safe通过读取options file的[mysqld\_safe]或[safe\_mysqld]部分启动了些安全功能，例如发生错误时重启服务器，将错误信息记录到错误日志。

取消查询:需要先输入相应的结束符，再输入\c。

## 数据类型

### 日期时间

```
  ts TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  dt DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
 ```

 上述两类数据类型可以在update时自动更新相应值，CURRENT_TIMESTAMP具有同意词(NOW(),LOCALTIME,LOCALTIME())。

## [mysql数据目录](https://dev.mysql.com/doc/refman/5.7/en/data-directory.html)

### [系统数据库](https://dev.mysql.com/doc/refman/5.7/en/system-schema.html)

mysql database存储着系统中重要的信息，这些存储信息的表按照授权系统，日志系统等类型进行分组。

## 日志

### 介绍

```
Error log				//Problems encountered starting, running, or stopping mysqld  
General query log		//Established client connections and statements received from clients  
Binary log				//Statements that change data (also used for replication)  
Relay log				//Data changes received from a replication source server  
Slow query log			//Queries that took more than long_query_time seconds to execute  
DDL log (metadata log)	//Metadata operations performed by DDL statements  
```

注意：1.如果开启了日志，一般日志存于mysql的data目录下。2.对于flush log，binary log将会关闭当前log file并根据index重新创建log打开，而err log,slow query log都只是close然后reopen。

### error log

通过配置log_error=file配置error log的输出，如果file不存在则拼接{host_name}.err在data目录下，如果都不存在则直接输出在console。

### general query log

对于一般查询日志会记录mysqld做了哪些事儿。会记录tcp/unix/管道等各种连接以及关闭，各种sql语句的执行，即使执行失败了。与binlog相比，query log会按照收到的语句顺序去记录而不是根据excuted but before any locks are released的顺序录入。(query log会记录查询语句，这些都是非常耗费性能的，所以默认是关闭的)
一般情况下query log是被禁止的。可以查看general_log系统变量，如果没有值或者1则表示query log是启用了的。如果为0则表示未启用，可以`set @@global.general_log = 1`启用。对于general_log_file=file_name，如果没有定义file_name仍然会拼接成host_name.log在data目录下。

### binary log

binlog主要包含描述数据库数据更改的statement,如表创建修改，表数据创建修改等操作。binlog在包含每个语句的同时还包含语句花费多少时间去更新的信息。binlog主要用于两个目的，一个是为了执行复制，源服务器发送包含在binlog的变化事件给副本以执行与源相同的操作，一个是为了进行恢复操作，当某一个时间点的数据被备份了，可以使用期间的binary log使得期间发生的命令继续再执行一次，以达到更新。当然binlog不会记录show,select等不会修改数据的语句。  
想要查看binlog相关参数可以使用`show variables like "%bin%"`。binlog的name由bin\_basename+.index组成，如果自己在basename里包含.extension，则extension会被忽略。
新的binlog的创建根据以下三种情况创建：1.服务器启动或重启。2.服务器flush了log。3.log文件的大小到达了max\_binlog\_size(binary log file有可能会大于max\_binlog\_size，因为有可能创建了一个比较大的事务，而事务内数据的写入是不会分开在其他文件中)。
二进制日志是一组文件，日志是一组二进制日志文件和一个索引文件组成。清空所有binlog可以使用reset master语句，但最好不要随意删除源上的旧二进制log文件。手动删除文件，最好采用purge binary logs来删除，它可以安全的更新索引文件。
purge binary logs例子：

```
purge binary logs to 'mysql-bin.010';                //指定清除文件
purge binary logs before '2019-04-02 22:46:26';      //根据日期来进行清除
```

对于before后面的datetime参数需保证'YYYY-MM-DD hh:mm:ss'的format。当副本正在进行复制时，此语句也可以安全运行，不需要额外阻止它。当活动副本正在读取尝试删除的日志文件之一，此语句不会删除正在使用的日志或该日志文件之后的日志文件，但会删除之前交早的日志文件。但如果副本读取时，碰巧已经删除了它未读取的日志文件，则该副本无法进行复制操作。
安全清理二进制日志文件，按照如下步骤执行：

1. 通过`show slave status`检查它正在读取哪个日志文件
2. 获取源服务器的binary log文件的列表，通过`show binary logs`。
3. 确定获取副本中最早的日志文件(应该是通过第一步确定，无多slaves此步骤暂时无法验证)。
4. 备份想要删除的日志文件。
5. 清除日志文件。

也可以设置`expire_logs_days`系统变量，使日志自动删除(感觉这样还是不好，应该是没有备份的。擦，一向什么都不配置的公司服务器居然配了10天，可能是避免太占空间了？也没见主从备啊，感觉主从备确实可以主定时删，从保留log即可)。还有一点需要注意的是如果你手动的使用rm语句删除了binary数据，再使用purge binary logs的语句将会报failed，这时需要edit ./[basename].index文件确保上面的文件确实存在。

binlog的写入会在任何锁被释放前或者commit之前，如对于非提交事务mysql在接受到commit指令之前会先缓存指令，然后在commit执行之前将记录写入binlog，而非事务表则会在执行完成后立即写入。当一个线程开始处理事务，会分配一个buffer(大小取决于binlog\_cache\_size)去缓存语句，如果语句的大小超过了buffer则会开一个临时文件去存储事务，临时文件在线程结束时会被删除。

对于binary log，row based logging并发的inser会被转换成为正式的插入如create ... select和inset ... select，这是为了还原备份操作时能确实重建表，而如果是statement-based logging就只会存储原始语句。

#### 异步和同步刷盘binlog

默认情况下sync\_binlog（同步写入）是开启的，这样能保证在事务commit之前能写入disk。但是这样会导致如果日志落盘和commit之间宕机，使得日志与事实不一的情况，mysql默认采用与redolog的xa事务解决此问题。在mysql崩溃重启后。在回滚事务之前mysql扫描最近的binary log文件检查最近事务的xid值以计算最近有效的binary log file(删除xid后的无效数据)，恢复正确的binlog日志。在某次事务中，由于两阶段提交的原理如果binlog能恢复成功，则就可以使用redolog恢复之前事务的执行，而binlog不能恢复成功，则本次事务无法恢复执行失败。(<s>我猜测外部程序员观察事务的执行情况应该是在binlog完成commit状态修改后，就会返回commit状态当然也有可能是真的完全执行完了再发送信息，发送信息过程也会出现丢失的情况，但对数据库而言就不关心这方面的具体内容了。</s>上述思考并不重要，mysql都宕机了还想api确定执行状态，直接查看mysql日志信息即可确认。)

#### binlog的格式

* --binlog-format=STATEMENT：基于sql语句的binlog格式
* --binlog-format=ROW：基于单个表的行是如何受影响的(注意，binlog本质上还是逻辑日志，row格式也是逻辑上的描述。)
* --binlog-format=MIXED：log的mode默认为statement，但是会基于一些情况转化成row

源服务器和副本使用不同的存储引擎存储数据是非常容易出现，对于这种情况基于statement的binlog容易出现非确定性的情况，mysql这时会将其标注为不可信赖且发布一个warning"Statement may not be safe to log in statement format."为了避免这种情况最好使用row-based的log格式。
对于运行时更改binlog的格式，如果采用session级别则可能有这么几种情况。1.session使用where更新很多很多匹配行，这样基于语句就比基于row要更加高效。2.花费了非常多的时间或者执行了很多的语句，但最终只改变了很少一部分行，这是采用row-based更加有效。
对于副本服务器，如果源服务器更改了log format，自己本身不会做相应转化，会执行报错处理。(这是不是说明，如果源服务器会更改的话，副本还是就采用mixed模式为好。)注意如果innodb采用提交读和读未提交，则只能使用row-based模式，哪怕在运行时更改了log模式会导致innodb不再具有插入的功能。
对于DML语言和DDL语言，DML会遵循binlog的格式而DDL则会保持statement的格式，如是混合型语言如`CREATE TABLE ... SELECT`，则DML部分和DDL部分分开对待。

### slow query log

慢查询日志记录超过long_query_time时间和min_examined_row_limit的行数的sql语句。通过slow_query_log开启关闭慢查询日志，通过slow_query_log_file制定文件名和路径，如果不指定则默认data目录进行{host_name}-slow.log拼接。为了包含DDL语句的慢查询，需要开启log_slow_admin_statements变量。进一步的，如果想包含不使用索引的查询，启动log_queries_not_using_indexes系统变量。

#### mysqldumpslow工具

调用格式为mysqldumpslow [options] [log_file]，一般来说mysqldumpslow会对结果进行分组，抽象string为s，number为n。

```
options说明：
-a				//不再将结果抽象
-g 				//匹配grep-style模式的结果
-h				//查看host_name的log
-l				//不减去整个时间的lock time
-r				//reverse
-s				//排序，t:查询时间，l:上锁时间，r:影响的行数,c:数量
-v				//verbose模式，展示更多的详细信息。
```

### redolog和binlog的区别

1. binlog产生于存储引擎的上层，不管什么存储引擎都会产生binlog，而redolog是在innodb层产生的。
2. binlog记录逻辑性语句，即便是基于行格式也是逻辑上的记录，如(表，行，修改前值，修改后值)，而redo log则是记录物理页上的修改，类似(pageId,offset,len,修改前值，修改后值)。
3. redolog是循环写，空间固定(只用记录最近的情况就行了)，而binlog是追加写。
4. 事务提交时，先写redolog写完后进入prepare状态，再写binlog写完后(末尾写入XID event表示写完)再进入commit状态(即在redo log里面写一个commit记录)。(两阶段提交是否会出现XID event写入但redo log commit写入失败的情况呢？应该是理论上仅存在微小可能，因为两者之间的执行间隔应该非常的短，可以忽略不计。)

## 配置管理

### 不同环境的启动配置

```
//可以使用1-2GB内存和拥有许多表，希望在中等数量的客户端上获得最大性能
key_buffer_size=384M
able_open_cache=4000
sort_buffer_size=4M
read_buffer_size=1M
//如果只有256M内存也只有几个表，但有很多排序
key_buffer_size=64m
sort_buffer_size=1m
//内存很少。连接很多
key_buffer_size =512k
sort_buffer_size=100k
read_buffer_size=100k
//甚至这样
key_buffer_size=512k
sort_buffer_size=16k
table_open_cache=32
read_buffer_size=8k
net_buffer_length=1k
//如果正在执行对比内存大很的表的group by或者order by操作，需要增大read_rnd_buffer_size加快排序操作后的行读取速度。
```

### [常用状态变量](https://dev.mysql.com/doc/refman/5.7/en/server-status-variables.html)（还有重要的系统变量等，有时间再细看）

使用`show global status`查看必要的状态变量;

```
aborted_clients								//由于客户端没有正确关闭连接的情况下死亡而中止的连接数
aborted_connects							//尝试连接mysql而失败的次数
```

### sql_mode

默认的sql模式包含如下几种**ONLY_FULL_GROUP_BY， STRICT_TRANS_TABLES， NO_ZERO_IN_DATE， NO_ZERO_DATE， ERROR_FOR_DIVISION_BY_ZERO， NO_AUTO_CREATE_USER，和 NO_ENGINE_SUBSTITUTION**。  
ONLY_FULL_GROUP_BY模式：默认开启，强制未分组字段必须使用聚合函数，如果关闭后不能进行分组的字段将会自由选取行数据以填充。
STRICT_TRANS_TABLES模式：和STRICT_ALL_TABLES模式一样属于严格sql模式，对于在事务表遇见错误都会回滚，对除以0，某些数据类型插入0值之类非法数值会报错。但对于非事务表的多个更新，插入操作出现错误，两者的区别在于本模式会警告并调整结果继续插入，而STRICT_ALL_TABLES则会出现错误中止操作导致部分更新。
 ERROR_FOR_DIVISION_BY_ZERO模式：未启用该模式则除以0会插入null且不警告，启用则插入null并警告，同时启用严格模式和本模式则会error。  

## 服务器

### 连接管理

 mysql处理与客户的连接分为3种类型：1.全平台可用的tcp/ip连接。2.Unix平台下Unix套接字连接。3.Windows平台下的共享内存连接请求，命名管道连接请求。（可以通过系统变量skip_networking，禁止tcp/ip连接，这通常用于仅本地使用的数据库）  
 连接线程的配置管理：1.thread_cache_size**状态变量**确定线程缓存大小，默认情况下服务器启动时自动调整，也可以在运行时设置。2.thread_stack设置服务器的线程栈空间的大小。3.threads_created为处理连接而创建的线程数，threads_cached线程缓存中的线程数。4.max_connections可以设置最大值，一般来说可以支持500到1000个连接，如果内存很大工作负载不高响应也快，最多可达10000个。其中最大连接连接值受文件描述符大小的硬性限制，包括open_files_limit**系统变量**，甚至操作系统文件描述符大小。

### 主机缓存

mysql只缓存非本机的TCP连接，对于环回地址，unix套接字，命名管道，共享内存建立的连接不使用缓存，缓存的表信息存于performance_schema数据库的host_cache表。主机缓存的目的在于：1.减少dns查询转换。(说明mysql为了方便管理员操作，默认使用域名而不是ip来标识，其中可以使用skip_name_resolve系统变量禁止dns解析)2.保存有关客户端连接过程中发生的错误信息（注意：系统变量host_cache_size控制缓存数据条数，如果超过了就会丢失最少使用的主机缓存，在丢失的过程中会损失错误信息，这样等新建立连接时如该主机连接是被阻止则阻止将会被放开。）。当然，如果因为网络故障导致连接被阻止，可以truncate或者正常sql操作处理host_cache表来满足需求。

## 数据备份和恢复

### mysqldump备份

常用dump指令：

```
mysqldump -uroot -p**** --all-databases >/tmp/all.sql                    //导出所有数据库数据
mysqldump -uroot -p**** --databases db1 --tables a >/tmp/a.sql     //导出a表数据，相关导出以此类推，注意：如果使用了--databases之类指令会默认创建db然后加载数据 ，而省略了的话可以导入到其他名字的db或者table上。
mysqldump -uroot -p**** db1 a where="id=1">/tmp/a.sql                   //条件导出db1.a表满足条件数据，注意密码一定要连着。
mysqldump --host=h1 -uroot -p**** --databases db1 | mysql --host=h2 -uroot -p*** db2   //跨服务器导入数据，db1数据库必须存在。
后缀参数说明：
--no-data					//只导出表结构
--no-create-info			//抑制住create相关的结构语句，只包含数据
--single-transaction			//该选项开启隔离级别为可重复读的事务，在dump data的时候。注意该指令只对innodb引擎的表生效，其他引擎的表依然会改变。该指令与lock-tables是隐性相冲突的，loca-tables直接就串行，不需要事务版本控制了。
--quick						//该选项用于dump检索行时一行一行的检索，而不是将整个行检索到内存中，然后再输出。
--lock-tables				//对于dumped的数据库，对所有的表添加读锁，以保证并发安全的dump数据库。但对于innodb表，最好使用single-transaction，节省性能。
--opt						//是一个组合指令包含--add-drop-table,--add-locks(insert优化有讲解),--disable-keys（只对mysiam有效，具体逻辑没看懂）,--quick,--lock-tables等一系列指令的集合。该选项在没有其他选项时是默认，需要使用--skip-opt去关闭相关指令。

源服务器复制副本
--master-data				//此选项在转储输出的时候包含一条change master to语句，该语句指示源服务器的二进制日志坐标（文件名和位置），使得副本服务器开始增量复制。
--dump-slave				//用于副本服务器转储，转储的master仍然是指向该副本的master而不是该副本。

使用分隔文本存储数据
--tab=/tmp							//使用已存在的目录存储文本

--fields-terminated-by=str			//用于分隔列值的字符串（默认值：tab）。
--fields-enclosed-by=char			//包含列值的字符（默认值：无字符）。
--fields-optionally-enclosed-by=char	//包含非数字列值的字符（默认值：无字符）。
--fields-escaped-by=char			//转义特殊字符的字符（默认：不转义）。
--lines-terminated-by=str			//行终止字符串（默认值：换行符）。

重新加载文本存储数据
mysql db1 < t1.sql						                //使用分隔文本存储时会默认输出包含建表的sql
mysqlimport db1 t1.txt							    //导入数据
mysqlimport --fields-terminated-by=,.... db1 t1.txt      //如果使用了其他数据格式
```

### 时间点恢复

#### binary log的时间点恢复

常用的检查log的指令:

```
show bianry logs;                //查看现有的bin log文件（虽然可以直接在data目录查看）
show master status;              //查看现在master数据库的状态，可以确定当前binary log文件名
mysqlbinlog binlog_files | mysql -u root -p  //直接将binlog输入mysql，倒入
mysqlbinlog binlog_files > tmpfile           //将log文件输出到文件中，并进行编辑
mysql -u root -p <tmpfile                    //再将编辑过的数据导入
mysqlbinlog binlog.01 binlog.02 | mysql -u root -p   //加载多个binlog应放在一起使用，如果分开使用将导致如第一个文件创建了临时表读完文件后就删除了，而读第二个文件却又依赖第一个表的问题。

mysqlbinlog binlog.000001 >  /tmp/statements.sql
mysqlbinlog binlog.000002 >> /tmp/statements.sql
mysql -u root -p -e "source /tmp/statements.sql"    //或者像这样，将文件转换成一个sql文件然后处理。
```

#### [使用事件位置的时间点恢复](https://dev.mysql.com/doc/refman/5.7/en/point-in-time-recovery-positions.html)

```
--start-datetime="2020-05-27 12:00:00"
--stop-datetime="2020-05-27 12:00:00"   
使用mysqlbinlog加上述参数去查看时间范围内的数据，同时使用grep -C筛选想要的数据，最后确定position行。

--start-position=1006
--stop-position=1868                    
输入使用增量日志例子:mysqlbinlog --start-position=1985 bin.123456| mysql -u root -p
```

### 备份恢复的逻辑

1.使用全量备份的sql导入数据库。2.使用增量备份的binlog，通过mysqlbinlog工具导入数据库。

## 优化

### 概述

数据库级别的优化

* 选择合适的表结构，每一列选择合适的数据类型，每一张表选择合适的列类型，如频繁更新数据的情况选择多表少行，分析大量数据的表选择多行少表。
* 建立合适的索引
* 选择合适的存储引擎，如事务存储选择innodb，非事务存储选择高拓展和高性能的mysiam。
* 每张表选择合适的行格式，如压缩表占用更少的硬盘空间和需要更少的I/O。innodb的各种workloads(啥意思？)均支持，mysiam只读表支持。
* 选择合适的锁策略
* 合适的使用缓存空间，足够大去存放频繁访问的数据，但又不会导致物理内存分页。该项配置包括innodb的buffer pool,mysiam的key cache和mysql的query cache。

硬件级别优化

* [磁盘寻道](https://www.cnblogs.com/jswang/p/9071847.html)的限制：现代磁盘的检索一般在10ms左右，因此我们一秒之内进行100次disk seeks。所以在一个磁盘内是很难去优化检索时间的，一般采用分布数据在多个磁盘内。
* 磁盘读写的限制：一个磁盘的传输速率在10-20MB/s，可以通过并行读磁盘去优化
* cpu周期：在主存中，我们需要处理数据，如果一个相比较内存而言比较大的表，这将是最主要的限制因素。
* 内存带宽：当cpu需要的数据超过了cpu缓存容量时，主存带宽就成为了瓶颈，这对大多数操作系统而言并不常见，但是需要注意

### select语句优化

`analyze table table_name`用于分析更新table的数据分布信息，如cardinality信息以便于优化器选择合适的优化策略。(因为table的数据分布在insert等update过程中，不可能实时的去更新表数据分布信息，这样会浪费很多性能，所以需要定期的更新下数据的分布信息。)

#### where子句优化

常规优化：1.去除不必要的括号。2.恒等式转化。3.恒等条件去除如1==1。4.索引使用的常量表达式只计算一次。5.检索前期就检测无效的表达式。6.having和where如果不使用聚合函数或group，则合并。7.优化连接的组合。8.使用覆盖索引。9.根据使用索引顺序进行排序。

#### 相关子查询

嵌套在其他查询中的查询称为子查询或内部查询，包含子查询的查询称为主查询或外部查询。
不相关子查询：内部查询独立于外部查询，内部查询仅执行一次，执行完毕后将结果作为外部查询的条件使用。
相关子查询：内部查询的执行依赖外部的查询数据，外部查询执行一次内部查询就会执行一次。即先从外部查询表中取出一个数据项，再将数据项传入内部查询执行内部查询，根据内部查询执行的结果判断。

#### range优化（注意：单个索引）

1. 对于rang优化mysql采用将各个区间转换为一个范围的方式进行优化，如重叠范围的条件将被合并，空范围的将被去掉。（注意like使用通配符作为前缀的匹配不会现阶段被使用,还有就是如果是等于之类的使用or,and才是range）
2. 对于索引多值的or/in匹配，mysql优化器会采用dive index的方式去准确评估索引每个区间的两端计算匹配的cost(也可以采用index statistics来估算，这样比较快但估算的准确度取决于统计信息的正确程度，可以使用analyze table去更新统计信息。)。当然value越多，dive index的时间越长，可以通过eq_range_index_limit系统变量去限制dive index的数量。不使用dive index有那么几种情况，1.存在子查询。2.使用聚合，group by。3.使用了force index。
3. 可以优化表达式如`SELECT ... FROM t1 WHERE ( col_1, col_2 ) IN (( 'a', 'b' ), ( 'c', 'd' ));`。但注意最好不要同时使用in，和or，如联合索引c1,c2,c3，c1 or in(c2,c3)只会使用c1索引而不会充分利用所有索引值。
4. range优化将会受到range_optimizer_max_mem_size系统变量的限制，如超出限制将会发出warning，并可能采用全表扫描。范围的内存的估计可以按一个or230字节，一个and125字节，如过是and和in的组合，in的每一个数据都按or来计算，然后两两相乘。

#### index merge

index merge通过对where后的多个不同索引条件下的range访问，并行使用索引查找，然后通过union,intersect,unions-of-intersections等方法进行合并得到一个结果。注意：满足条件后，mysql也不是一定会这样操作，具体看explain。  
index merge的使用条件：1.对于条件的n parts的等于表达式，索引确实能完全包含表达式中的所有字段。2.对主键使用了范围查询的条件。(依赖索引的filter，如果filter<60%直接忽略该索引，进行全表scan。)  

如果有比较复杂and/or/in语句，mysql的优化可能有些限制需要使用如下分配律对式子进行修改(应该可使用explain检查是否使用了index merge，或者测试性能时间)：

```
(x AND y) OR z => (x OR z) AND (y OR z)
(x OR y) AND z => (x AND z) OR (y AND z)
```

可以在explain中查看extra字段来判断index merge，展示结果如Using intersect(...)，Using union(...)，Using sort_union(...)

此部分的控制可在**optimizer_switch**里面查看，如index_merge,index_merge_intersection,...等参数状况。启动单一算法，可以关闭index_merge，然后开启单一算法。

intersection：一般and条件拼接的使用intersect拼接，其中有三点需要注意。1.如果多个索引在进行并行查找时，已满足查询所需字段的要求，则不会回表去查询full row，此时可以在extra 查看有标明using index。2.如果索引并未完全覆盖查询，则会先完成索引条件相关的查询，然后再查询**full row**。3.如果查询条件包括主键，则主键并不会进入查询条件，而是会对查询结果进行过滤。但其实对于intersect方法的拼接有更好的处理方法，比如根据and条件建立联合索引。  
union：一般对于or进行union拼接，对于包含and的语句，逻辑为先intersection再进行union，上述的分配律规则也是在揭示这一点，主要应该还是为了减少计算吧。
sort-union：处理对于union不适用的>,<等范围情况，该查询会在返回结果前对结果进行排序。例子如下，包括范围查询以及索引不覆盖所有条件。

```
SELECT * FROM tbl_name
  WHERE key_col1 < 10 OR key_col2 < 20;

SELECT * FROM tbl_name
  WHERE (key_col1 > 10 OR key_col2 = 20) AND nonkey_col = 30;
```

#### 索引下推

索引下推是对使用索引的sql语句的一种优化，当使用了索引下推后，explain将会在extra字段展示using index condition语句。索引下推直接将符合索引判断的where条件都下放到存储引擎，使得能够完成联合索引的情况，也可以实现一次性in的多值判断的情况。这样一方面存储引擎可以减少访问全行数据，减少基表查数据的次数，同时也减少了mysql服务器访问存储引擎的次数。可通过`SET optimizer_switch = 'index_condition_pushdown=on';`指令来启动索引下沉。using index condition 和 using index的区别：主要在于索引下沉需要访问full row，而覆盖索引在索引上就能的到想要的结果不用访问基表。

#### 多范围读取优化

使用二级索引的范围扫描取行可能会导致对基表的多次随机磁盘访问，而磁盘多范围扫描优化(multi-rang read)则是通过索引收集的键对其进行排序，减少随机访问磁盘次数。该项配置有optimizer_switch系统变量的mrr(启用优化)，mrr_cost_based(是否不偏向使用优化，因为存在如覆盖索引是访问索引完后不会再访问基表，这个排序操作没什么意义等情况。)其中用于排序索引的缓冲区大小由read_rnd_buffer_size系统变量控制。

#### 条件过滤

在表连接时，启用条件过滤可以使前缀表在选择索引时能不仅仅考虑where后面的索引条件，会考虑额外的索引外的条件。比如如果a索引能检索回1000行，而b索引能检索回10000行但能使用等值的条件过滤1%变成100行返回给后续的表使用，这时应该选择b索引（注意：1.条件只能是常量。2.条件过滤的where条件不能在索引内）。控制条件过滤开关的系统变量："condition\_fanout\_filter"。（注意：在连接时，filtered的统计结果在最后不需要输出时，就没再统计了，所以是100%）

#### order by优化

index:为了避免filesort的额外开销，mysql可能会采用index来进行排序，即使index并没有完全匹配上order by后面的条件。如果使用select *去order by index，大概率不会使用索引，因为虽然可以通过索引去排序，但是排完序去源表拿完整行数据会导致多次随机I/O明显代价高于全表扫描然后排序，当然，如果只用select index倒是会使用索引（其实就是如果不用回表就会使用索引）。对于存在where key_part1 order by key_part2的情况，通过key_part1索引访问的行都是根据part2来排序的，通过这个结果也是可以节约排序操作的，对于随机访问I/O次数的问题select条件也许会筛除掉不少row减少代价，两者代价的取舍看mysql分析的统计情况了。使用group by会产生隐式排序，但请不要依赖隐式排序，请显式使用order by虽然会被优化器优化。  
filesort: 内存中的排序，buffer的大小依赖于`sort_buffer_size`的大小，且必须容纳15个元组。每个元组的大小由`max_sort_length`系统变量配置，每行数据超过该系统变量部分则默认为相等。可以通过`show global status;`查看排序信息，如`Sort_scan`变量可以查看全表扫描排序的次数，`Sort_merge_passes`表示排序算法merge passes(指合并)出现次数。关于外部排序算法的选用，通过`max_length_for_sort_data`设置的行数值来进行调控，如果结果超过该值则使用算法1，未超过则使用算法3。(如果将该值设置的过高，则会导致磁盘I/O过高，这是由于导致临时文件过大，排序操作导致过多I/O。)

msyql排序模式：

1. <sort_key,rowid>:根据索引或者全表扫描按照过滤条件将查询到的排序所需字段值和rowID组成键值对，存入sort buffer中。如果sort buffer内存大于这些键值对所需内存，则不需要创建临时文件，否则每次使用快排在内存中排好序写入临时文件中。最后将这些临时文件使用磁盘外部排序，将row id写入到结果文件中。为了保证回表时减少随机I/O的次数可以将rowid排序后再回表(通过`read_rnd_buffer_size`系统变量控制内存大小)。
2. <sort_key,additional_fields>:与1的区别在于第一次查询时就将所需数据查完组成键值对，这样就可以避免排序完成后的回表查询。（与1相比多占用空间，节约了时间）
3. <sort_key,packed_additonal_fields>:对varchar进行了压缩，类似压缩链表吧。

外部排序：
1.mysql的外部排序采用的是多路归并(7路)的方式进行的，通过多次归并减少文件数量最后得到一个。2.mysql的临时文件只有一个，通过位偏移量来区分多个。3.对于limit则是采用优先队列的方式进行淘汰的，如果是limit m,n则一般是存m+n最后丢弃m的方式来实现。如果是存在外部排序的话，则还是采用原外部排序的方法，然后取相应位置的值即可。

临时文件：
对于外部排序所需要的临时文件，可以查看`tmpdir`系统变量来获取临时文件位置。对于Unix系统使用:对路径进行分割，建议使用多个物理磁盘而不是同一磁盘的多个分区进行分割。

#### group by优化

group by使用三种方式来实现，其中前两种会利用索引信息。

1. loose index scan实现：仅通过索引信息即可扫描出结果，且不必详细比对所有索引，仅判断部分索引值便可得出结果。extra会出现using index for group-by
2. tigh index scan实现：仅凭索引扫描出结果分组，但缺少比对信息，需要查询出相应数据与where比对后过滤输出。
3. 建立临时表：全表扫描，建立临时表分组。extra会出现Using temporary。

### 子查询优化

对于子查询mysql根据不同情况采用不同策略进行优化。  
对于IN的情况：Semijoin，Materialization，EXISTS strategy
对于NOT IN的情况：Materialization, EXISTS strategy

select * from supplier_cdn_vod_202106 as s1 join cdn_vod_202106 as s2  on s1.cdn = s2.cdn and s1.domain = s2.domain and s1.country = s2.country and s1.time = s2.time and s1.type = s2.type and s1.type = "month";

### 索引优化

#### 概述

索引虽然大大增加了查询速度，但是为不必要的列添加索引会造成如下缺点：1.索引会占用一定的空间2.mysql会花时间去抉择使用index。3.在增删改时会造成额外的代价去更新索引。

#### 索引失效

1. 不满足最左匹配原则
2. filter条件不满足(mysql5.7居然没有？还是没开启详细信息之类的配置)
3. like使用%前缀，或非常量
4. 在使用or时如果or分隔的and组不存在索引，则不会使用索引(and优先级高于or)
5. order by的失效情况

#### 索引拓展

innoDB通过将每个二级索引附加主键列来拓展二级索引，即将二级索引和主键一起组合成联合索引，可以通过`SET optimizer_switch = 'use_index_extensions=off';`开启。可以通过explain查看是否采用索引拓展的区别，主要可以通过rows查看索引查到的行数对比，或者ref能看明显用了2个const，或者key_len长度变化，或者没有使用using where表示没有在服务器用where条件筛选。([using where存疑,啥情况下会出现using where?吗的，使用主键查本地是正常的using index，服务器却会using where using index，不晓得为啥](https://dev.mysql.com/doc/refman/5.7/en/explain-output.html#explain_extra))

## innodb引擎

## ACID

### 一致性

一致性体现在mysql崩溃时对数据的安全的保护，主要包含innodb双写buffer和崩溃恢复。
双写buffer：innodb的pagesize一般为16kb，而文件系统对数据页的写入并不是原子操作，需要将多个数据页分别写入磁盘。而如果系统在这个时候崩溃，就会导致数据的不一致问题。其中系统崩溃分为两种情况，一种是物理页完整即不存在数据页只写了一半的问题，这只需要通过redo log进行崩溃恢复即可。而第二种情况是有些数据页有一部分写入成功导致partial page write问题，其中redo log记录的是对页的物理修改，无法对已损坏了的数据页进行redo操作(因为写了一半的页无法直接处理，如遇见删除第十行数据的操作，该操作不具备幂等性，且不知道执行到哪一步了。)。此时mysql的双写策略就派上了用场。mysql在写入数据时先将脏页的修改写入redo log文件，然后通过两阶段提交完成binlog和redo log写入，最后状态位置为commit.然后再将将数据拷贝入double buffer，double buffer再先写入共享表空间(ibdata文件，但需要注意的是8.0.20之前位于系统表空间，之后则有专门的双写文件，可通过innodb\_doublewrite\_dir，innodb\_doublewrite\_files等参数调控)，再将数据写入真实的数据文件。此时虽然双写磁盘会造成额外的开销，但是开销基本上都会远远小于两倍，因为共享表空间的数据是连续存放顺序写入的，而真实的数据文件则是随机I/O写入的。  
所以mysql对此类情况的处理是在恢复时先检查page的checksum,checksum就是检查page的最后事务号，已被损坏的页是无法通过校验的。如果未通过校验，mysql先找到系统表空间的该页副本直接替换来进行还原.而如果是完整的数据页，则判断先系统表空间是否有副本，即双写文件又没写好，如果没有则通过redo log来重做页数据。

完整恢复流程：1.表空间发现。2.应用redo log。3.回滚未完成事务。4.change buffer合并。5.清除。

### 事务隔离级别

一般默认情况是可重复读，但对于大批量数据报告，对数据的精度和重复性要求低于锁的开销，建议使用低级别的读已提交和读未提交的隔离级别。而串行化则一般用于XA事务，或者定位并发和死锁问题。

* 读未提交:锁的使用类似读已提交，select读取是最新的快照。
* 读已提交:每次select读都是已提交的最新快照。
* 可重复读:为了确保可重复性，每次select时只读取第一次的快照，不会受先开启事务但查看前修改的会话影响(除非自己修改)。其他与读已提交差距参考锁章节即可或者[官网](https://dev.mysql.com/doc/refman/5.7/en/innodb-transaction-isolation-levels.html)。
* 串行读:类似于可重复读，但是隐式将所有的简单select语句转化成in share mode。如果自动提交开启，则为每个select都生成一个事务。

## [InnoDB的锁](https://zhuanlan.zhihu.com/p/149228460)

* 共享(s)和独占锁(x):也就是读写锁，持有读锁的可以共享，读写，写写互斥。
* 意向锁:意向锁是表级锁(mysql支持多粒度的锁，如表锁和行锁)，表明该事务随后将在该表的哪一行使用共享锁(is)或独占锁(ix)(意思是设置或者说比较时是表锁，用起来是行锁？)，意向锁的主要目的就是用作行锁，或者表明某人将要使用行锁。如`lock tables ... write`将对表设置x，`select ... lock in share mode`将对表设置is,而`select ... for update`将设置ix。意向锁的主要目的是加快mysql自己的锁冲突判断速率，如A表已先被某行加入行锁了，另一个事务想对表加入表级写锁，这是如果没有意向锁存在将会一行一行的遍历记录查看是否有行锁存在。表级锁的兼容性如下：

||X|IX|S|IS|
|----|----|----|----|----|
|X|Conflict|Conflict|Conflict|Conflict|
|IX|Conflict|Compatible|Conflict|Compatible|
|S|Conflict|Conflict|Compatible|Compatible|
|IS|Conflict|Compatible|Compatible|Compatible|

* 记录锁:记录锁用于利用索引锁住所有相关修改，如`SELECT c1 FROM t WHERE c1 = 10 FOR UPDATE`将阻止t.c1=10的所有行insert,delete和update。1. 不匹配行的记录锁将会在mysql评估完where条件后释放，所谓不匹配行记录主要是如果没有走索引将会通过扫描的方式去扫记录。2. for update语句如果遇见已经加锁的情况，将会给出最新已提交版本(半一致性读),以保证where的条件能正确筛选更新。(上述两点均只发生在读已提交隔离级别，对于第二点，在获取最新commited值后如果确实要更新就必须去获取锁了。)所以行锁在InnoDB中其实是基于索引实现的，当扫描到该表的索引，将会对该索引设置锁，所以一旦某个加锁操作没有使用索引，如果是可重复读的隔离级别根据二阶段锁协议那么该锁就会大概率**退化为表锁**。
* 间隙锁(gap):用来锁住一段范围的索引记录防止插入，如`SELECT c1 FROM t WHERE c1 BETWEEN 10 and 20 FOR UPDATE`将会锁住10-20的范围，阻止15的插入无论是否已经存在了该值。间隙锁对于提交读及以下的隔离级别是不会生效的，而以上的隔离级别也仅对对非唯一索引和多值唯一索引生效。间隙锁可以跨越单个索引，多个索引甚至是空(指不存在的值)，是并发性和性能之间的折中。不同事务的间隙锁是允许重叠的，两个间隙锁的范围将会被合并(取消被两个事务中较晚的那个？)。因为间隙锁是单纯的抑制类锁，重叠后造成的效果是一致的。当然，可以通过使用读已提交的隔离级别和开启`innodb_locks_unsafe_for_binlog`配置，使间隙锁仅仅在外键约束和重复检查这两个地方生效。
* Next-Key锁:next-key锁是记录锁和索引上间隙的组合，这里会锁住两边的间隙和该条记录。因为非唯一索引可能会对索引记录插入重复值，所以next-key会锁住该记录的间隙且包含记录本身。对于防止两个事务同时插入，可以使用共享锁锁定范围，引发next-key锁锁定。这里解释下插入不是有互斥锁嘛，为什么不能用互斥锁来防止插入。记录锁是不会锁不存在的记录的，加锁逻辑我觉得可能是在插入后，如果同时插入的话，记录锁是不能生效的。
* 插入意向锁(IIgap):插入意向锁是行级锁，主要与间隙锁冲突，而插入意向锁如果值不冲突那么锁也是不冲突的。(在完成间隙锁的锁定后，锁定范围内使用插入意向锁会冲突，这样在一定范围内不会出现幻读。)对于先使用插入意向锁的插入，next-key锁定范围将以之前插入意向锁的记录截止。

||record|gap|next-key|IIgap|
|----|----|----|----|----|
|record|Conflict|Compatible|Compatible|Compatible|
|gap|Compatible|Compatible|Compatible|Conflict|
|next-key|Compatible|Compatible|Compatible|Conflict|
|IIgap|Compatible|Compatible|Compatible|Compatible|

* [自增锁](https://dev.mysql.com/doc/refman/5.7/en/innodb-auto-increment-handling.html):自增锁是特殊的表锁，对于性能和并发性需要根据算法来均衡。

## MVCC(多版本并发控制)

InnoDB是一个多版本的存储引擎，保留已更改行的旧版本信息以支持事务功能，使其能支持事务并发和回滚的特性，这些信息存在系统表空间或者undo表空间。InnoDB通过这些数据，去执行事务的回滚操作，当然也可以据此去构建更早期的行版本以实现一致性读(实现非加锁版本的一致性读)。  
InnoDB为表的每一行内部增加了三个字段分别为:

1. 6byte大小DB_TRX_ID字段，表示最后一个事务的事务标示(在行插入，更新时更新，其中删除被认为是更新，专门使用1byte表示deleted)。
2. 7byte的DB_ROLL_PTR回滚指针字段，该指针指向被写入回滚段的undo log。如果该行更新了，undo log将记录必要的可以恢复更新前记录的信息。
3. 6byte的DB_ROW_ID单调递增行id，如果InnoDB的聚簇索引是自动生成的(应该是指创建时主键成为聚簇索引)，那么该索引将包含row id值，否则row id将不出现在任何索引中。

undo logs在回滚段被分为插入和更新log，insert log仅用在事务回滚或提交后立马删除，而update log用在一致性读。而update undo log只有在没有事务存在时才会被丢弃，因为InnoDB此时已经为其分配了一个快照了(指bin log吧，redo log迟早会被删的)。

官方建议定期提交事务，包括一致性读的事务，否则InnoDB将无法丢弃undo log导致回滚段过大占满其所在的表空间，为了方便管理建议使用undo表空间。在InnoDB多版本控制的情况下，InnoDb并不会立即物理上删除行信息，只会在记录了删除的update undo log被丢弃时才执行删除(此时删除的操作被称为purge,清除)。

## 索引与MVCC

InnoDB的MVCC在对待聚簇索引和二级索引是有区别的，聚簇索引的更新是本地更新，且有隐藏行指针指向undo log(此处维护跟表row一致，有很多隐藏字段。或者说，InnoDB的表结构就是按照聚簇索引来构建的，所以结构理所应当一样。)，而二级索引这些都是没有的。所以在使用二级索引时，如果发现该记录已经被标记删除或者被新事务更新，则需要走聚簇索引去得到undo log然后据此完成版本复原获取所需值。

其中在二级索引中，对于覆盖索引和索引下推的处理方式是不同的。如果是在事务中有值被更新或者删除了，索引覆盖是不会被启用的，而索引下推会将所有符合条件的值全部返回回来。然后，将返回回来的值通过聚簇索引校验比对版本，不符合要求的删除或修复。(对于此段，有两个问题。1.覆盖索引不能启用，那么检测更新或删除是如何做的，难道是一旦开启了事务就不能使用覆盖索引了？2.全部查过来，是如何获取对应值的所有版本号的，难道是直接查log?不然感觉会有遗漏，比如当前是但被改过了，就不是了，然后查不到。)

## 非锁一致性读

非锁的一致性读意味着InnoDB使用多版本并发控制去读数据库在某时刻的快照。所谓一致性读，就是该查询只能看见在本事务开始时间点之前提交的事务，不能看见之后的或未提交的事务更改。但这个规则的例外就是，同一个事务的查询能否看见自己之前的修改。这个意外对应一种异常情况，即本事务之前修改过一个行，但由于这个行没有提交，从规则上来讲应该看不见自己的修改。如果有其他会话同时也修改了该行，你甚至有可能会看见表中数据处于一种在数据库中从没有存在过的状态。(A事务查看r1->B事务查看r1->B事务更新r1+1->A事务更新r1+2->A事务提交->B事务查看r1状态?)

如果事务是默认的隔离级别--可重复读，所有的一致性读将会读在本事务第一次快照读建立时的值。如果需要读取更新的快照，需要结束本事务然后开启新的查询。而如果是读已提交的隔离级别，则会读取该数据最新的快照。(实际上学到这，应该就能明白依赖快照纯mvcc并不能完全解决并发性问题，还是得依赖锁之类的工具去解决。举个例子，两个事务并行执行先查库存大于1，然后减1的情况。如果两个事务同时查到还剩最后一个，然后同时减一，mvcc并不能阻止bug的产生。)一致性读是读已提交和可重复读隔离级别的默认读取模式，但如果是select ... for update则将被mysql认为是DML，一致性读失效。在mysql中，DML如update,delete,insert等语句一致性读是失效,取而代之的是使用加锁来处理，如下例子：

```mysql
SELECT COUNT(c1) FROM t1 WHERE c1 = 'xyz';
-- Returns 0: no rows match.
DELETE FROM t1 WHERE c1 = 'xyz';
-- Deletes several rows recently committed by other transaction.

SELECT COUNT(c2) FROM t1 WHERE c2 = 'abc';
-- Returns 0: no rows match.
UPDATE t1 SET c2 = 'cba' WHERE c2 = 'abc';
-- Affects 10 rows: another txn just committed 10 rows with 'abc' values.
SELECT COUNT(c2) FROM t1 WHERE c2 = 'cba';
-- Returns 10: this txn can now see the rows it just updated.
```

一致性读对于某些DDL语言也是失效的，如DROP_TABLE和ALTER_TABLE复制源表然后删除源表等操作。然后就是对于INSERT INTO ... SELECT,UPDATE ... (SELECT)等字句中并未定义FOR UPDATE和LOCK IN SHARE MODE的情况会根据配置不同产生不同差异。如果是默认情况下将会使用锁，然后SELECT部分将会采用读已提交的模式读取最新副本。当然如果使用`innodb_locks_unsafe_for_binlog`模式，将会避免使用锁。

## 读锁

### 注意

1. 一致性读会忽略锁的存在，因为锁并不能锁旧版本。
2. 对于子查询，子查询必须自己也使用锁，不然外部查询并不会对子查询的数据加锁。

### 实践

* 对于查询是否存在然后再更改删除的操作，建议使用`lock in share mode`读锁锁住不让其他事务同时操作(for update也行，如果要更改确实必须互斥锁，但是一开始就使用代价相比较大)，然后再执行。
* 对于排号计数的查询，不应该使用一致性读或者共享读，这样会导致有可能两个人的号是一样的。而且如果先使用共享锁，再尝试更新时有可能会导致死锁，(由于二阶段锁协议，只有在事务commit时才释放锁，所以会拿了读锁后继续拿写锁。)所以建议直接使用for update直接拿排他锁。(或者使用last_insert_id()函数，细节以后再看)

## 幻读

## [undo log](http://mysql.taobao.org/monthly/2015/04/01/)

一份undo log日志是由单个读写事务的undo log记录集合组成。其中undo log记录包含了如何撤销上个事务的更改对聚簇索引记录的更改。undo log存储在全局临时表空间(临时表空间存储InnoDB非压缩的临时表和相关对象，在MYSQL5.7版本被引入。可以通过`innodb_temp_data_file_path`配置临时表的路径，name等临时表空间相关属性。如果该项配置未配置，则会创建一个12mb的自拓展名为ibtmp1的文件，在data目录。临时表空间文件在mysql启动时被创建--如未创建成功则无法启动，关闭或退出时被删除，如果是崩溃退出则不会被删除，可手动删除或mysql保持相同配置重启时重建。)中，他们不会被redo-logged，因为他们不需要在崩溃时进行回复工作。

每个undo表空间或者全局临时表空间最大支持128个回滚段，当然也可以通过`innodb_rollback_segments`来定义回滚段的数量。而每个回滚段支持多少事务，取决于回滚段的undo slots数量和每个事务需要多少undo logs。回滚段undo slots的数量取决于InnoDB页大小(由下表可知槽大小为16b)，如下所示:
|InnoDB 页大小|每个回滚段的Undo Slots大小(page size/16)|
|:----:|:----:|
|4KB|256|
|8KB|512|
|16KB|1024|
|32KB|2048|
|64KB|4096|
每个事务最多被分配4个undo logs，其中每个对应以下操作类型：

1. 用户定义的表中的插入操作。
2. 用户定义的表中的更新|删除操作。
3. 用户定义的临时表中的插入操作。
4. 用户定义的临时表中的更新|删除操作。

Innodb只会按需分配所需的undo log，对于用户对普通表的操作将会把undo log分配至系统表空间或则undo表空间的回滚段，对于临时表的操作会将undo log分配至临时表空间的回滚段。由之前的数据信息可以分析出同时执行的事务并发上线，当undo slots不够用时，该事务将会重新尝试执行。当然，如果事务操作发生在临时表，则事务的读写并发同时受限于临时表空间的32大小的回滚段。

```mysql
(page_size/16)*32       //临时表只有插入或update操作,如果同时存在，按上文分析则会对半分，数量下降一半。
(innodb_page_size/16)*(innodb_rollback_segements)    //普通表只有插入或update操作，如果同时存在，下降一半。

```

## redo log

redolog在磁盘上的物理表示为ib_logfile0和ib_logfile1两个文件,mysql通过循环写的方式写入redo log。redo log的数据段，通过一直增加的LSN值进行表示。(LSN,log sequence number,一个永远增长64bit的数字，对应表示redo log中操作记录的时间点。这个时间点与事务边界无关，可以掉落在一个或多个事务的中间。它不止可以用在崩溃恢复，也可以用来管理buffer pool。)

### redo log的大小和数量修改

1. 正确关闭mysql
2. 编辑conf中的`innodb_log_file_size`和`innodb_log_files_in_group`分别修改大小和数量。
3. 重启mysql

在这里，InnoDB如果检测到配置与实际不符合，会写入一个log checkpoint，然后关闭和移除老的log文件，创建新的log并打开。

### redo log的组提交刷新

InnoDB像大多数服从ACID的数据库一样，刷新redolog在事务committed之前。但为了提高吞吐量，InnoDB避免每次commit就刷新日志，而是将同时刻的多个事务一起commit。

## 死锁

### 死锁的预防

1. 尽量使用事务的行锁而不是表锁。
2. 操作数据库表数据时，尽量按照统一的顺序。
3. 尽量对索引使用行锁，不然会退化成表锁。
4. 事务尽量的小，且快速提交
5. 由于两阶段锁，所以读锁升级写锁很容易造成死锁。

### 处理死锁

mysql自动检测死锁，如果产生了死锁，mysql将会选择代价相对较低的事务进行回滚，事务的代价是根据insert,update,delelte行数决定的。当然通过`innodb_lock_wait_timeout`设置超时回滚，处理死锁有时可能更有效，当由于死锁等待的事务过多时。

死锁状态确认:`show engine innodb status`;SHOW ENGINE INNODB STATUS 和 InnoDB monitor都是监控InnoDB状态的手段，需要熟练使用。

## mysql物理结构

![mysql架构](https://dev.mysql.com/doc/refman/8.0/en/images/innodb-architecture.png)

* 聚簇索引:innoDB存储引擎存储的表都是有聚簇索引的，一般为主键，如果没有主键则寻找合适的非空unique索引，如果还是没有则生成包含行id值的合成列作为隐藏聚簇索引，长度为6字节。  
* 二级索引:聚簇索引外的称之为二级索引，当聚簇索引建立后，所有二级索引的叶子结点都为聚簇索引的key。所以如果索引列(主键也一样)很长，索引就会占用更多的空间，同时意味着更多的I/O，对查询也是不利的。  
* 索引的物理结构:InnoDB索引都是采用B树结构，特殊的多维数据空间索引采用R树。索引页的大小默认为16k，由innodb_page_size配置项决定。
* 索引插入策略:当新纪录插入innoDB索引中时，InnoDB将尝试保留1/16分之一的页面空闲空间以供将来的插入和更新(感觉应该是用来避免b+树的多次分裂合并操作)。一般顺序批量插入是1/16，随机插入则是1/2到1/16不等。

### 排序索引构建

InnoDB创建或重建索引时执行批量加载而不是一次一条的插入，这种创建方法称为排序索引构建，空间索引不支持排序索引构建。索引构建分为三个阶段，第一阶段扫描聚簇索引生成索引条目并添加到排序缓冲区，直到排序缓冲区变慢，缓冲区数据输出到临时中间文件。第二阶段将这些中间文件数据进行归并排序。第三阶段将临时文件的数据插入到B树中。  
普通插入批量数据的弊端:在一次一条的插入中先从根节点往下寻找插入位置，先使用乐观插入，如因页面已满导致插入失败时会执行悲观插入，这将导致拆分和合并B树节点。而批量插入时可以一次性构建必要深度的B树结构，从底至上构建，批量进行插入避免拆分和合并。
为了将来的索引增长留出空间，可以使用innodb_fill_factor配置项来设置B树页面保留空间的百分比，例如innodb_fill_factor=80将会保留20%的空间。当然此配置不会对text or blob条目生效，保留空间量也有可能和配置不完全相同，因为该配置被解释为提示而不是硬限制。

## 表空间

### 系统表空间

系统表空间是用于存储InnoDB数据字典，双写缓冲区，变化缓冲区以及undo log的地方。其中，当表创建在系统表空间而不是在表文件或者通用的表空间时，还有会存储存储表数据和索引。默认情况下，在mysql的data目录中创建一个名为ibdata1的系统表空间数据文件。
调整系统表空间大小:`innodb_data_file_path=ibdata1:10M:autoextend`,其中三个参数分别代表数据文件名，初始大小以及自动增长属性。自动增长大小可以通过`select @@innodb_autoextend_increment;`查看，一般增量为8M。以下情况说明当dir为空字符串时可以定义完整路径在path。(注意系统表空间数据文件可以不止一个，如:`innodb_data_file_path=ibdata1:50M;ibdata2:50M:autoextend`)
>
[mysqld]  
innodb_data_home_dir =  
innodb_data_file_path=/myibdata/ibdata1:50M:autoextend

## Buffer Pool

### instance

buffer pool的实例大小为innodb\_buffer\_pool\_size/innodb\_buffer\_pool\_instances两个系统变量的计算，每个instance都有自己的锁，信号量，物理块(buffer chunks)以及逻辑链表。每个instance之间是相互独立的，可以并发读写，在数据库启动时被分配，在数据库关闭内存时释放。(当innodb\_buffer\_pool\_size小于1GB时，instances被重置为1，主要防止太多小的instance影响性能。)

### buffer chunks

buffer chunks是最底层的物理块，由两部分组成:1.控制体和与其对应的数据页。控制体中包含指针指向数据页，数据页包含控制信息如行锁，自适应hash和用户存储的信息。

### 逻辑链表

链表节点就是数据页的控制体，各种类型链表的节点拥有相同属性，方便管理。

1. Free List:上面的节点均为未被使用的节点，Innodb需要保证Free List有足够的节点提供给用户线程使用，否则从FLU List或LRU List淘汰一定的节点。
2. LRU List:LRU List按照最少使用算法排序，由两部分组成默认前5/8为young list存储经常被使用的热点page，后3/8为old list。新读入的page默认加到old list头，只有满足一定条件后才被移到young list上，主要是为了预读数据页和防止全表扫描污染buffer pool。
3. FLU List:该链表上都是脏页节点，FLU List上的页面一定存在在LRU List上。由于数据页可能会在不同时刻被修改多次，数据页上记录了最老的一次修改的lsn，FLU List的节点按照oldest_modification
排序，链表的尾端是最早被修改的数据页。(FLU List通过flush_list_mutex保证并发)

### Buffer Pool预热

MYSQL在重启时缓冲池没有什么数据，需要业务对数据库进行数据操作才能慢慢填充。所以在初期MySQL的性能不会特别好，特别Buffer Pool越大预热过程越长。为了缩短预热过程，可以把之前Buffer Pool中的页面数据存储到磁盘，等MySQL启动时直接加载磁盘数据即可。其中dump过程就是将数据页按space_id,page_no组成64位数字，写到外部文件中。

## change buffer

change buffer主要用于管理不在buffer pool的二级索引相关的更改。主要二级索引不像聚簇索引一样插入相对集中且是非唯一的，修改二级索引相关页内容将会导致大量随机I/O影响性能，所以需要集中缓存二级索引相关修改然后随后合并多次修改再写入磁盘中。change buffer的合并写入可能会持续几个小时，此时将会明显影响磁盘相关查询的I/O。当mysql关闭时，内存中的change buffer内容将会存入磁盘上系统表空间中的change buffer空间。

## [optimizer_trace](https://www.imooc.com/article/308721)

## 分库分表

### 使用符号连接

使用datadir符号链接，将表空间分区分散到多个磁盘增加效率，可使用show variables like 'datadir'查看。

### 分区

注意事项：1.做分区时要么不定义主键，要么把分区字段加入到主键中。2.分区字段不为null。

### 难点

* 分布式事务的问题
* 跨节点Join的问题
* 跨节点合并排序分页的问题
* 多数据源管理问题

CREATE TABLE ti (id INT, amount DECIMAL(7,2), tr_date DATE)
    ENGINE=INNODB
    PARTITION BY HASH( MONTH(tr_date) )
    PARTITIONS 6;

## mycat

[相关文档1](https://www.yuque.com/books/share/6606b3b6-3365-4187-94c4-e51116894695/fb2285b811138a442eb850f0127d7ea3)
[相关文档2](http://www.mycat.org.cn/document/mycat-definitive-guide.pdf)