# [mysql](https://dev.mysql.com/doc/refman/5.7/en/innodb-deadlocks-handling.html)
## 基础
mysql的模式匹配:`like`语句支持%以匹配一个或多个,`_`匹配一个，也可以通过`regexp`来使用正则表达式。为使得`regexp`区分大小写可以使用`binary`转化成二进制字符串，如`regexp binary '^b'`。(^:以...开头，$:以...结尾,.:任意一个字符,{n}:前一个规则重复5次。)  
mysqld_safe:在UNIX上启动mysqld服务器的推荐方法，mysqld\_safe通过读取options file的[mysqld_safe]或[safe_mysqld]部分启动了些安全功能，例如发生错误时重启服务器，将错误信息记录到错误日志。

取消查询:需要先输入相应的结束符，再输入\c。
## [mysql数据目录](https://dev.mysql.com/doc/refman/5.7/en/data-directory.html)
### [系统数据库](https://dev.mysql.com/doc/refman/5.7/en/system-schema.html)
mysql database存储着系统中重要的信息，这些存储信息的表按照授权系统，日志系统等类型进行分组。
## 日志
### 介绍
Error log				//Problems encountered starting, running, or stopping mysqld  
General query log		//Established client connections and statements received from clients  
Binary log				//Statements that change data (also used for replication)  
Relay log				//Data changes received from a replication source server  
Slow query log			//Queries that took more than long_query_time seconds to execute  
DDL log (metadata log)	//Metadata operations performed by DDL statements  

注意：1.如果开启了日志，一般日志存于mysql的data目录下。2.对于flush log，binary log将会关闭当前log file并根据index重新创建log打开，而err log,slow query log都只是close然后reopen。
### error log
通过配置log_error=file配置error log的输出，如果file不存在则拼接{host_name}.err在data目录下，如果都不存在则直接输出在console。
### general query log
对于一般查询日志会记录mysqld做了哪些事儿。会记录tcp/unix/管道等各种连接以及关闭，各种sql语句的执行，即使执行失败了。与binlog相比，query log会按照收到的语句顺序去记录而不是根据excuted but before any locks are released的顺序录入。(query log会记录查询语句，这些都是非常耗费性能的，所以默认是关闭的)
一般情况下query log是被禁止的。可以查看general_log系统变量，如果没有值或者1则表示query log是启用了的。如果为0则表示未启用，可以`set @@global.general_log = 1`启用。对于general_log_file=file_name，如果没有定义file_name仍然会拼接成host_name.log在data目录下。
### binary log
binlog主要包含描述数据库数据更改的statement,如表创建修改，表数据创建修改等操作。binlog在包含每个语句的同时还包含语句花费多少时间去更新的信息。binlog主要用于两个目的，一个是为了执行复制，源服务器发送包含在binlog的变化事件给副本以执行与源相同的操作，一个是为了进行恢复操作，当某一个时间点的数据被备份了，可以使用期间的binary log使得期间发生的命令继续再执行一次，以达到更新。当然binlog不会记录show,select等不会修改数据的语句。  
想要查看binlog相关参数可以使用`show variables like "%bin%"`。binlog的name由bin_basename+.index组成，如果自己在basename里包含.extension，则extension会被忽略。
新的binlog的创建根据以下三种情况创建：1.服务器启动或重启。2.服务器flush了log。3.log文件的大小到达了max_binlog_size(binary log file有可能会大于max_binlog_size，因为有可能创建了一个比较大的事务，而事务内数据的写入是不会分开在其他文件中)。
二进制日志是一组文件，日志是一组二进制日志文件和一个索引文件组成。清空所有binlog可以使用reset master语句，但最好不要随意删除源上的旧二进制log文件。手动删除文件，最好采用purge binary logs来删除，它可以安全的更新索引文件。
purge binary logs例子：

```
purge binary logs to 'mysql-bin.010';				      //指定清除文件
purge binary logs before '2019-04-02 22:46:26';	      //根据日期来进行清除
```
对于before后面的datetime参数需保证'YYYY-MM-DD hh:mm:ss'的format。当副本正在进行复制时，此语句也可以安全运行，不需要额外阻止它。当活动副本正在读取尝试删除的日志文件之一，此语句不会删除正在使用的日志或该日志文件之后的日志文件，但会删除之前交早的日志文件。但如果副本读取时，碰巧已经删除了它未读取的日志文件，则该副本无法进行复制操作。
安全清理二进制日志文件，按照如下步骤执行：

1. 通过`show slave status`检查它正在读取哪个日志文件
2. 获取源服务器的binary log文件的列表，通过`show binary logs`。
3. 确定获取副本中最早的日志文件(应该是通过第一步确定，无多slaves此步骤暂时无法验证)。
4. 备份想要删除的日志文件。
5. 清除日志文件。

也可以设置`expire_logs_days`系统变量，使日志自动删除(感觉这样还是不好，应该是没有备份的。擦，一向什么都不配置的公司服务器居然配了10天，可能是避免太占空间了？也没见主从备啊，感觉主从备确实可以主定时删，从保留log即可)。还有一点需要注意的是如果你手动的使用rm语句删除了binary数据，再使用purge binary logs的语句将会报failed，这时需要edit ./[basename].index文件确保上面的文件确实存在。

binlog的写入会在任何锁被释放前或者commit之前，如对于非提交事务mysql在接受到commit指令之前会先缓存指令，然后在commit执行之前将记录写入binlog，而非事务表则会在执行完成后立即写入。当一个线程开始处理事务，会分配一个buffer(大小取决于binlog_cache_size)去缓存语句，如果语句的大小超过了buffer则会开一个临时文件去存储事务，临时文件在线程结束时会被删除。

对于binary log，row based logging并发的inser会被转换成为正式的插入如create ... select和inset ... select，这是为了还原备份操作时能确实重建表，而如果是statement-based logging就只会存储原始语句。
#### 异步和同步刷盘binlog
默认情况下sync_binlog（同步写入）是开启的，这样能保证在事务commit之前能写入disk。但是这样会导致如果日志落盘和commit之间宕机，使得日志与事实不一的情况，mysql默认采用xa事务解决此问题。在mysql崩溃重启后。在回滚事务之前mysql扫描最近的binary log文件检查最近事务的xid值以计算最近有效的binary log file，恢复正确的日志。
#### binlog的格式
* --binlog-format=STATEMENT：基于sql语句的binlog格式
* --binlog-format=ROW：基于单个表的行是如何受影响的
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
show bianry logs;							//查看现有的bin log文件（虽然可以直接在data目录查看）
show master status;							//查看现在master数据库的状态，可以确定当前binary log文件名
mysqlbinlog binlog_files | mysql -u root -p		//直接将binlog输入mysql，倒入
mysqlbinlog binlog_files > tmpfile				//将log文件输出到文件中，并进行编辑
mysql -u root -p <tmpfile					//再将编辑过的数据导入
mysql binlog.01 binlog.02 | mysql -u root -p	//加载多个binlog应放在一起使用，如果分开使用将导致如第一个文件创建了临时表读完文件后就删除了，而读第二个文件却又依赖第一个表的问题。
```
#### 使用事件位置的时间点恢复

```
--start-datetime="2020-05-27 12:00:00"
--stop-datetime="2020-05-27 12:00:00"		//使用上述参数去过滤时间范围的log
--start-position=1006
--stop-position=1868						//不建议使用datetime去输出log，因为这样会有遗漏的风险，常用datetime过滤范围，根据具体的内容找到相应postion进行输出。输出格式同上，只需要加上后缀即可。
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
在表连接时，启用条件过滤可以使前缀表在选择索引时能不仅仅考虑where后面的索引条件，会考虑额外的索引外的条件。比如如果a索引能检索回1000行，而b索引能检索回10000行但能使用等值的条件过滤1%变成100行返回给后续的表使用，这时应该选择b索引（注意：1.条件只能是常量。2.条件过滤的where条件不能在索引内）。控制条件索引的系统变量："condition\_fanout\_filter"。（注意：在连接时，filtered的统计结果在最后不需要输出时，就没再统计了，所以是100%）
#### order by优化
index:为了避免filesort的额外开销，mysql可能会采用index来进行排序，即使index并没有完全匹配上order by后面的条件。如果使用select *去order by index，大概率不会使用索引，因为虽然可以通过索引去排序，但是排完序去源表拿完整行数据会导致多次随机I/O明显代价高于全表扫描然后排序，当然，如果只用select index倒是会使用索引（其实就是如果不用回表就会使用索引）。对于存在where key_part1 order by key_part2的情况，通过key_part1索引访问的行都是根据part2来排序的，通过这个结果也是可以节约排序操作的，对于随机访问I/O次数的问题select条件也许会筛除掉不少row减少代价，两者代价的取舍看mysql分析的统计情况了。使用group by会产生隐式排序，但请不要依赖隐式排序，请显式使用order by虽然会被优化器优化。  
filesort: 内存中的排序，buffer的大小依赖于`sort_buffer_size`的大小，且必须容纳15个元组。每个元组的大小由`max_sort_length`系统变量配置，每行数据超过该系统变量部分则默认为相等。可以通过`show global status;`查看排序信息，如`Sort_scan`变量可以查看全表扫描排序的次数，`Sort_merge_passes`表示排序算法merge passes(指合并)出现次数。

msyql排序模式：

1. <sort_key,rowid>:根据索引或者全表扫描按照过滤条件将查询到的排序所需字段值和rowID组成键值对，存入sort buffer中。如果sort buffer内存大于这些键值对所需内存，则不需要创建临时文件，否则每次使用快排在内存中排好序写入临时文件中。最后将这些临时文件使用磁盘外部排序，将row id写入到结果文件中。为了保证回表时减少随机I/O的次数可以将rowid排序后再回表(通过`read_rnd_buffer_size`系统变量控制内存大小)。
2. <sort_key,additional_fields>:与1的区别在于第一次查询时就将所需数据查完组成键值对，这样就可以避免排序完成后的回表查询。（与1相比多占用空间，节约了时间）
3. <sort_key,packed_additonal_fields>:对varchar进行了压缩，类似压缩链表吧。

外部排序：
1.mysql的外部排序采用的是多路归并(7路)的方式进行的，通过多次归并减少文件数量最后得到一个。2.mysql的临时文件只有一个，通过位偏移量来区分多个。3.对于limit则是采用优先队列的方式进行淘汰的，如果是limit m,n则一般是存m+n最后丢弃m的方式来实现。如果是存在外部排序的话，则还是采用原外部排序的方法，然后取相应位置的值即可。

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
**Loose Index Scan access**
## mysql物理结构
* 聚簇索引:innoDB存储引擎存储的表都是有聚簇索引的，一般为主键，如果没有主键则寻找合适的非空unique索引，如果还是没有则生成包含行id值的合成列作为隐藏聚簇索引，长度为6字节。  
* 二级索引:聚簇索引外的称之为二级索引，当聚簇索引建立后，所有二级索引的叶子结点都为聚簇索引的key。所以如果索引列(主键也一样)很长，索引就会占用更多的空间，同时意味着更多的I/O，对查询也是不利的。  
* 索引的物理结构:InnoDB索引都是采用B树结构，特殊的多维数据空间索引采用R树。索引页的大小默认为16k，由innodb_page_size配置项决定。
* 索引插入策略:当新纪录插入innoDB索引中时，InnoDB将尝试保留1/16分之一的页面空闲空间以供将来的插入和更新(感觉应该是用来避免b+树的多次分裂合并操作)。一般顺序批量插入是1/16，随机插入则是1/2到1/16不等。

### 排序索引构建
InnoDB创建或重建索引时执行批量加载而不是一次一条的插入，这种创建方法称为排序索引构建，空间不支持排序索引构建。索引构建分为三个阶段，第一阶段扫描聚簇索引生成索引条目并添加到排序缓冲区，直到排序缓冲区变慢，缓冲区数据输出到临时中间文件。第二阶段将这些中间文件数据进行归并排序。第三阶段将临时文件的数据插入到B树中。  
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


## 死锁
### 处理死锁
死锁状态确认:`show engine innodb status`;
