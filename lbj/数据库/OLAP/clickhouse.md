# clickhouse

## 部署启动

### clickhouse下载

官网下载[clickhouse](https://clickhouse.com/docs/zh/getting-started/install),其中mac版使用`curl -O 'https://builds.clickhouse.com/master/macos/clickhouse' && chmod a+x ./clickhouse`指令下载可执行文件。

### 初始启动

该可执行文件通过`./clickhouse server`和`./clickhouse client`来启动。由于这样启动是默认default账户，在使用账户启动时需要下载[server.xml](https://github.com/ClickHouse/ClickHouse/blob/master/programs/server/config.xml)和user.xml。server.xml的必要参数看这个[链接](https://zhuanlan.zhihu.com/p/470885530)
正常起的clickhouse-server指令:`./clickhouse server --config-file=${filepath}`,clickhouse-client指令:`clickhouse-client -u ${user} --password ${password}`。

admin账号的创建如下：

```sql
CREATE USER IF NOT EXISTS admin IDENTIFIED WITH sha256_password BY '${password}'
GRANT ALL ON *.* TO admin WITH GRANT OPTION
```

## [备份](https://aop.pub/artical/database/clickhouse/backup-recovery/)

### 规模较小时

```shell
//使用dump的方式进行备份
clickhouse-client -d database --query="SELECT * FROM [db.]tablename format CSV" > export_tablename.csv

//进行插入工作
cat export_tablename.csv | clickhouse-client --query="INSERT INTO [db.]tablename FORMAT CSV"; //半角逗号（’,’）分割符

clickhouse-client --query "INSERT INTO tutorial.hits_v1 FORMAT TSV" --max_insert_block_size=100000 < hits_v1.tsv //制表符（tab,’t’）
```

## 特点

适用场景：宽表查询，且每列点数据较小入数字和短字符串(URL这种，60字节左右)。每个查询只有一个大表，除了它外，其他都很小。主要任务就是使用原始数据在线提供各种数据报告。

优点：

1. 紧凑的数据：列式数据库管理系统应该除了数据本身外不存储其他额外的数据，即采用紧凑的方式存储数据而不应该在值旁边存储它们的长度，如果不是这样将对cpu的使用产生强烈影响。
2. 数据压缩：除了选择在磁盘空间和CPU消耗之间进行不同权衡的高效通用压缩编解码器，还针对特定类型数据的专用编解码器使得ClickHouse能与更小的数据库(时间序列数据库)的竞争中超越它们。
3. 数据的磁盘存储：许多列式数据库只能在内存中工作，这种方式会造成比实际更多的设备预算。ClickHouse被设计工作在传统磁盘上，提供更低的存储成本，但如果可以使用SSD和内存它也会合理利用这些资源。
4. 分布式处理：一般的列式数据库不支持分布式的查询处理。(跨实例分析会比较困难？)ClickHouse支持将数据保存在不同的shard上，每个shard都有容错备份的replica组成，查询并行的在shard上进行对用户是透明的。使用异步的多主复制技术，保证系统在不同副本上保持相同的数据。
5. 向量引擎与并发处理：数据不仅仅按列存储，同时还按向量(列的一部分)进行处理，这样更加高效的使用CPU，如列a的8维向量与列b的向量进行join、加减分析操作。所以ck与传统关系型数据库不同在于并行查询，且默认单查询cpu使用数为服务器核数的一半。
6. 实时数据更新、查询和索引：支持在线查询，低延时响应，不用提前离线计算缓存。数据可以持续不断高效的写入到表中，但写入的过程中不会存在任何加锁行为。可以定义主键索引，为使利用主键的快速进行范围查找，数据总是以增量的方式有序的存储在MergeTree中。
7. 支持近似计算：用于近似计算的各类聚合函数，如：`distinct value`,`medians`,`quantiles`，以加快速度，如果需要精确计算则使用精确版本。方式：1.如基于数据的部分样本进行抽样近似查询。2.不使用全部的聚合条件，随机选择有限个聚合条件进行聚合，这在某些分布条件下提供准确的聚合结果又同时降低了计算资源的使用。
8. 适应性连接算法：支持自定义join多个表，ck倾向使用散列连接算法，但如果有多个大表则会采用合并-连接算法。

缺点：

1. 没有完整事务支持。
2. 缺少高频率，低延迟的修改删除已存在数据的能力，仅能用于批量删除或修改数据。
3. 稀疏索引使得ClickHouse不适合通过其键检索单行的点查询，即应对小批量高并发的场景。

其他：

1. 不支持prepared queries

## 性能

吞吐量：可以使用每秒处理的行数或每秒处理的字节数来衡量。如果数据放置在page cache中，则一个不复杂查询在单个服务器上能够以2-10GB/s的速度处理，如果是简单查询可以达到30GB/s(未压缩的数据)。如果数据没在page cache中，则速度取决于压缩率，一般磁盘允许400MB/s的速度读取速度。对于分布式处理，处理速度是线性扩展的，受限于聚合或排序的结果不是那么大(这块内容还是不太理解，为什么受限？)。(对于传统数据库的落盘一般是directIO，由自己控制刷盘逻辑防止丢失数据。读的话我感觉应该还是走操作系统那套缓存流程，目前没发现有什么特殊需求。)

短查询延迟时间：普通查询使用主键且没有太多行(几十行)进行处理，没有查询太多列，数据有被page cache缓存延迟应该小于50ms(最优的情况下小于10ms)。如果使用的HDD(机械硬盘)，在数据没有预加载的情况下，查询所需的时间延迟可通过一下公式计算出:`查找时间(10ms)*查询列数*查询的数据块的数量`(所以数据块数量是什么意思？)

处理大量短查询的吞吐量：一般情况下ClickHouse可以在单台服务器上每秒处理数百个查询，但这不适合此数据库的业务场景，建议每秒查询次数不超过100。

数据的写入性能：建议每次写入不少于1k行的批量写入或每秒不超过1个写入请求。当使用tab-separated格式写入MergeTree表中时，写入速度大约为50-200MB/s。为了提高写入性能，也可以使用并行insert数据。(写应该是直写)

## 配置

### 查看配置

```sql
SELECT name, value, changed, description
FROM system.settings -- 配置的系统表
WHERE name LIKE '%max_insert_b%'
FORMAT TSV -- 制表符tab分隔

max_insert_block_size    1048576    0    "The maximum block size for insertion, if we control the creation of blocks for insertion."
```

## 分区

```sql
-- 查看分区信息
SELECT
    partition,
    name,
    active
FROM system.parts
WHERE table ='bw_nerve_new'
-- 201901 是分区名称。1 是数据块的最小编号。3 是数据块的最大编号。1 是块级别（即在由块组成的合并树中，该块在树中的深度）。
```

## [相关SQL语法](https://clickhouse.tech/docs/zh/sql-reference/statements/create/)

### <div id = 120.1>稀疏索引及mysql</div>

密集索引：文件中的每个搜索码值都对应一个索引值(一对一)
稀疏：索引文件只为索引码的某些值建立索引(一对多，因为是按顺序排列的所以可以按区间建立索引)
myisam的所有索引都是稀疏索引，innodb有且只有一个密集索引，非聚簇索引都是稀疏索引。

[clickhosue性能优化](https://huaweicloud.csdn.net/6335739dd3efff3090b57420.html)
[修改配置强行解决内存爆掉的问题](https://blog.csdn.net/anyitian/article/details/115390396)
[clickhouse踩坑](https://blog.csdn.net/qq_42016966/article/details/110487663)
[ck调研](https://xie.infoq.cn/article/9f325fb7ddc5d12362f4c88a8)
[ck分析研究好文](https://www.cnblogs.com/traditional/p/15218743.html)