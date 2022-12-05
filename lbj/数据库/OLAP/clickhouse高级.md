# clickhouse高级

## 数据库引擎

默认使用Atomic数据库引擎，提供可配置的表引擎和sql dialect

### atomic

支持非阻塞的drop table和rename table以及原子的exchange tables t1 and t2操作。

创建语句：

```sql
CREATE DATABASE test[ ENGINE = Atomic];
```

数据库Atomic中的所有表都有唯一的UUID，并将数据存储在目录/clickhouse_path/store/xxx/xxxyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy/，其中xxxyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy是该表的UUID。其中rename操作不会等待使用表的查询完成，而是立即执行，因为rename查询是在不更改uuid和移动表数据的情况下执行的。

drop/detach tables操作不删除任何数据，只是将元数据移动到/clickhouse_path/metadata_dropped/标记为已删除，并通知后台线程。最终表数据删除前的延迟由database_atomic_delay_before_drop_table_sec设置指定。可以使用sync修饰符指定同步模式，通过database_atomic_wait_for_drop_and_detach_synchronously配置完成。

```sql
RENAME TABLE new_table TO tmp, old_table TO new_table, tmp TO old_table; -- 非原子操作
EXCHANGE TABLES new_table AND old_table; -- 原子操作
```

对于ReplicatedMergeTree表，建议不要在ZooKeeper和副本名称中指定engine-path的参数。在这种情况下，将使用配置的参数default_replica_path和default_replica_name。

### MaterializedMySQL

该引擎具有实验性的特性，不应该在生产中使用。clickhouse服务器作为mysql副本工作，读取binlog并执行DDL和DML查询。

创建语句：

```sql
CREATE DATABASE [IF NOT EXISTS] db_name [ON CLUSTER cluster]
ENGINE = MaterializeMySQL('host:port', ['database' | database], 'user', 'password') [SETTINGS ...]
```

## 表引擎

### MergeTree

MergeTree引擎被设计用于插入极大量的数据到一张表中，插入数据将会一部分一部分的插入，并按照规则在后台进行合并。这种方式比在插入时不断的重写已存储的数据要更高效

特点：

1. 数据按主键存储，这将创建一个小的稀疏索引去更快的查找数据。(所谓索引就是按字段建立顺序，方便算法查找)
2. 支持分区
3. 支持副本
4. 支持数据采样方法

建表语句：

```sql
CREATE TABLE [IF NOT EXISTS] [db.]table_name [ON CLUSTER cluster]
(
    name1 [type1] [DEFAULT|MATERIALIZED|ALIAS expr1] [TTL expr1],
    name2 [type2] [DEFAULT|MATERIALIZED|ALIAS expr2] [TTL expr2],
    ...
    INDEX index_name1 expr1 TYPE type1(...) GRANULARITY value1, -- 跳数索引，按照granularity的value1的数量，组合成一个大块写入索引信息
    INDEX index_name2 expr2 TYPE type2(...) GRANULARITY value2, -- type2表示索引类型，如minmax,set(max_rows)(max_rows代表最多保留那么多个去重值)，bloom_filter
    ...
    PROJECTION projection_name_1 (SELECT <COLUMN LIST EXPR> [GROUP BY] [ORDER BY]),
    PROJECTION projection_name_2 (SELECT <COLUMN LIST EXPR> [GROUP BY] [ORDER BY])
) ENGINE = MergeTree() -- 指定引擎，其中mergetree()没有参数
ORDER BY expr -- 一个列组成的元组(如：order by (CounterID,EventDate))或者任意表达式，如果没有指定primary key，ck将会使用sort key作为主键。
              -- 如果不需要排序，可以使用order by tuple()
[PARTITION BY expr] -- 大多数情况下不需要分区键，分区并不会加快查询。不要使用过细粒度的分区键，不要使用分区字段，而是将分区字段标识为order by 表达
                    -- 式的第一列来指定分区。按月分区可以使用表达式toYYYYMM(date_column),这里的data_column是一个date类型的列。分区名的格式会是"YYYYMM"
[PRIMARY KEY expr] -- 如果选择与排序键不同的主键，在这里指定，大部分情况下不需要专门设置。
[SAMPLE BY expr] -- 用于抽样的表达式，如果要用抽样表达式，主键必须包含这个表达式如:SAMPLE BY intHash32(UserID) ORDER BY (CounterID,EventDate, intHash32(UserID))
[TTL expr -- 指定行存储的存储时间，并定义数据段在盘或者卷上的移动规则。表达式中必须至少存在一个Date或者DateTime类型的列，如：TTL date + INTERVAl 1 DAY。
    [DELETE|TO DISK 'xxx'|TO VOLUME 'xxx' [, TO DISK 'xx'... ] -- 默认移除规则是delete，可以移动到指定存储器，也可以有多个移动规则。
    [WHERE conditions]
    [GROUP BY key_expr [SET v1 = aggr_func(v1) [, v2 = aggr_func(v2) ...]]]]
[SETTINGS name=value, ...] -- mergetree行为的额外参数，如index_granularity —— 标记索引的间隔，默认8192，storage_policy —— 存储策略
```

主键与排序键:

主键列的数量没有明确限制，按需求可以在主键包含数量不等的列(类似联合索引？)这样可以：

1. 改善索引的性能
2. 改善数据压缩，ck以按主键排序数据段，数据一致性越好压缩越好。
3. 在CollapsingMergeTree 和 SummingMergeTree 引擎里进行数据合并时的处理逻辑依赖主键，此时指定与主键不同的排序键也是有意义的

如果不想使用主键，可以使用order by tuple。如果此时在使用插入时希望保证数据排序，使用max_insert_threads = 1。同样的按存储顺序查询，需要使用单线程查询。ClickHouse不要求主键唯一，所以您可以插入多条具有相同主键的行。

Clickhouse可以做到指定一个跟排序键不一样的主键，此时排序键用于在数据段中进行排序，主键用于在索引文件中进行标记的写入。这种情况下，主键表达式元组必须是排序键表达式元组的前缀(即主键为(a,b)，排序列必须为(a,b,**))。这是由于 SummingMergeTree 和 AggregatingMergeTree 会对排序键相同的行进行聚合，所以把所有的维度放进排序键是很自然的做法。但这将导致排序键中包含大量的列，并且排序键会伴随着新添加的维度不断的更新。在这种情况下合理的做法是，只保留少量的列在主键当中用于提升扫描效率，将维度列添加到排序键中。
排序键修改操作`alter table {table} modify order by (c1,c2)`(only columns added by the ADD COLUMN command in the same ALTER query, without default column value)，MergeTree排序键的修改只能是在原有的基础上添加新增的字段，或者在原有的基础上删除后面的字段，可以使用新增列到排序键中，但不能使用已有数据的列。所以对排序键进行 ALTER 是轻量级的操作，因为当一个新列同时被加入到表里和排序键里时，已存在的数据片段并不需要修改。由于旧的排序键是新排序键的前缀，并且新添加的列中没有数据，因此在表修改时的数据对于新旧的排序键来说都是有序的。

TTL:

TTL表达式的计算结果必须是日期或者日期时间类型的字段`TTL time_column，TTL time_column + interval`，如`TTL date_time + INTERVAL 1 MONTH,TTL date_time + INTERVAL 15 HOUR`。列过期后，ck会替换成为该列的默认值，如果列所有值都过期了，ck会删除此列。注意：TTL字句不能单独用于主键字段，此时建议直接在表语句上让该"行"过期。

```sql
-- 表过期转移删除
ORDER BY d
TTL d + INTERVAL 1 MONTH [DELETE],
    d + INTERVAL 1 WEEK TO VOLUME 'aaa',
    d + INTERVAL 2 WEEK TO DISK 'bbb';

-- 修改表TTL
ALTER TABLE example_table
    MODIFY TTL d + INTERVAL 1 DAY;
-- 修改列TTL
ALTER TABLE example_table
    MODIFY COLUMN
    c String TTL d + INTERVAL 1 MONTH;
```

ClickHouse 在数据片段合并时会删除掉过期的数据。当ClickHouse发现数据过期时, 它将会执行一个计划外的合并。要控制这类合并的频率, 可以设置 merge_with_ttl_timeout。如果该值被设置的太低, 它将引发大量计划外的合并，这可能会消耗大量资源。如果在两次合并的时间间隔中执行 SELECT 查询, 则可能会得到过期的数据。为了避免这种情况，可以在SELECT之前使用OPTIMIZE(线上不要使用，手动合并消耗性能，容易发生意外)。

数据存储：

表由按主键排序的数据段组成，当数据插入到表中时，会创建多个数据段并按主键的字典序排序。如主键是(CounterID, Date) 时，片段中数据首先按 CounterID 排序，具有相同 CounterID 的部分按 Date 排序。不同分区的数据会呗分成不同的数据段，ck在后台进行数据的合并，合并机制不保证具有相同主键的行都合并在同一个数据段中。数据段以wide或compact格式存储，wide格式每一个列都会在文件系统中存储为单独的文件，compact则是所有列都存储在一个文件中，compact格式可以提高插入量小插入频率高时的性能(可能是单文件的I/O会比较快)。数据存储格式由min_bytes_for_wide_part和min_rows_for_wide_part表引擎参数控制。如果数据片段中的字节数或行数少于相应的设置值，数据片段会以compact格式存储，否则会以wide格式存储。
granules是ck中数据查询时的逻辑上的最小不可分割数据集，包含整数个行。granules的第一行使用主键值进行标记，额外的索引文件会记录标记号和主键值。使用稀疏索引查找时，每个数据块中最多会多读index_granularity * 2行额外的数据。

表目录下的文件：
columns.txt：列名以及数据类型
count.txt：记录数据的总行数
primary.idx：主键索引文件，用于存放稀疏索引的数据。通过查询条件与稀疏索引能够快速的过滤无用的数据，减少需要加载的数据量。
{column}.bin：列数据的存储文件，以列名+bin为文件名，默认设置采用 lz4 压缩格式。
{column}.mrk2：列数据的标记信息，记录了数据块在 bin 文件中的偏移量。标记文件首先与列数据的存储文件对齐，记录了某个压缩块在 bin 文件中的相对位置；其次与索引文件对齐，记录了稀疏索引对应数据在列存储文件中的位置。clickhouse 将首先通过索引文件定位到标记信息，再根据标记信息直接从.bin 数据文件中读取数据

### ReplacingMergeTree

该引擎和 MergeTree 的不同之处在于它会删除排序键值相同的重复项，但是不能完全依赖，因为只会在数据合并期间进行去重。

建表语句：

```sql
ENGINE = ReplacingMergeTree([ver]) -- ver版本列,类型为UInt*,Date或DateTime。如果ver列指定，则保留ver最大的行，不然就保留最后一行。
```

### SummingMergeTree

该引擎继承自MergeTree。区别在于，当合并SummingMergeTree表的数据片段时。ClickHouse会把所有具有相同排序键的行合并为一行，该行包含了被合并的行中具有数值数据类型的列的汇总值。如果主键的组合方式使得单个键值对应于大量的行，则可以显著的减少存储空间并加快数据查询的速度。推荐将该引擎和MergeTree一起使用，例如，在准备做报告的时候，将完整的数据存储在MergeTree表中，并且使用SummingMergeTree来存储聚合数据。这种方法可以使你避免因为使用不正确的主键组合方式而丢失有价值的数据。(为何不自己写聚合函数，然后将结果存于mergetree?难道是因为排序键会经常更改？)

建表语句：

```sql
ENGINE = SummingMergeTree([columns]) -- columns列元组，包含了将要被汇总的列的列名，所选列必须是数值类型切不可被包含进主键。
```

注意：ClickHouse会按段合并数据，以至于不同的数据段中会包含具有相同主键的行，即单个汇总段将会是不完整的。因此，聚合函数sum()和GROUP BY子句应该在SELECT查询语句中被使用。

### AggregatingMergeTree

SummingMergeTree的升级版，并改变了数据片段的合并逻辑。ClickHouse会将一个数据片段内所有具有相同主键(准确的说是排序键,即排序键主键分离)的行替换成一行，这一行会存储一系列聚合函数的状态。AggregatingMergeTree具有和SummingMergeTree一致的特性，不能完全依赖其聚合结果，需要在查询时使用聚合函数保证聚合结果正确。

建表语句：

```sql
... (
    name1  AggregateFunction(func, type), -- AggregateFunction是特殊的类型，func指定聚合函数如uniq,quantiles，type指定存储类型
)
ENGINE = AggregatingMergeTree()
...
```

该引擎一般用作物化视图加快查询效率，注意插入规则需要按func+'State'规则使用，查询需要func+'Merge'规则使用，如下所示：

```sql
-- 创建实时跟踪原表的物化视图
CREATE MATERIALIZED VIEW test.basic
ENGINE = AggregatingMergeTree() PARTITION BY toYYYYMM(StartDate) ORDER BY (CounterID, StartDate)
AS SELECT
    CounterID,
    StartDate,
    sumState(Sign)    AS Visits, -- 满足func+'-State'规则
    uniqState(UserID) AS Users
FROM test.visits
GROUP BY CounterID, StartDate;

-- 聚合查询
SELECT
    StartDate,
    sumMerge(Visits) AS Visits, -- 满足func+'Merge'规则
    uniqMerge(Users) AS Users
FROM test.basic
GROUP BY StartDate -- 防止聚合结果未合并，得到准确结果
ORDER BY StartDate;
```

### CollapsingMergeTree

存在经常变化！！！的数据时，保存一行记录并在其发生任何变化时更新记录是合乎逻辑的，但是更新操作对 DBMS 来说是昂贵且缓慢的，因为它需要重写存储中的数据。此时可以使用该引擎，在写入行时加入特定列Sign。并且需要两个及以上的INSERT请求来创建不同的数据片段，如果我们使用一个请求插入数据，ClickHouse 只会创建一个数据片段是不会执行任何合并操作，也就不会执行折叠操作了。注意：只允许严格连续插入，即有单线程使用限制。


建表语句：

```sql
... (
    ...
    UserID UInt64,
    PageViews UInt8,
    Duration UInt8,
    Sign Int8 -- 特殊行标记折叠
)
ENGINE = CollapsingMergeTree(Sign)
...
```

安全的查询语句：

```sql
SELECT
    UserID,
    sum(PageViews * Sign) AS PageViews, -- 乘以系数以抵消
    sum(Duration * Sign) AS Duration
FROM UAct
GROUP BY UserID
HAVING sum(Sign) > 0 -- 去除等于0的以删除行，不然会得到0值的重复结果
```

### VersionedCollapsingMergeTree

 VersionedCollapsingMergeTree用于和CollapsingMergeTree相同的目的，但使用不同的折叠算法，允许以多个线程的任何顺序插入数据。主要是通过version列帮助正确折叠，即使在多线程的情况下，以错误的顺序插入。

 建表语句：

```sql
... (
    ...
    UserID UInt64,
    PageViews UInt8,
    Duration UInt8,
    Sign Int8, -- 特殊行标记折叠
    Version UInt8  -- 版本号，版本相同的才会折叠，用于应对多线程乱序插入
)
VersionedCollapsingMergeTree(sign, version)
...
```

安全的查询语句：

```sql
SELECT
    UserID,
    sum(PageViews * Sign) AS PageViews, -- 乘以系数以抵消
    sum(Duration * Sign) AS Duration
    Version
FROM UAct
GROUP BY UserID, Version -- 以版本号区分，如果存在多个版本号，可能会有多个数据
HAVING sum(Sign) > 0 -- 去除等于0的以删除行，不然会得到0值的重复结果。或许可以加一个Version = max(Version)的筛选条件
```

## 物化视图

物化视图将视图结果转储到实体表中，这样可以对实体表建立索引，进行修改维护也比较方便。且根据之前测试，聚合mergetree引擎对物化视图在join时，存在一定的问题，case如下：

```sql
CREATE MATERIALIZED VIEW pili_hera.test_wide_bw_03 to bw_nerve_wide_o1
AS SELECT
    a.en_name as en_name,
    a.real_server_id as real_server_id,
    a.record_time as record_time,
    a.in_bw as in_bw,
    a.out_bw as out_bw,
    a.max_bw as max_bw,
    isp_map.isp_id as isp_id,
    isp_map.isp_name as isp_name
FROM 
    (SELECT 
        en_name,
        real_server_id,
        record_time,
        isp,
        avgMerge(in_bw) AS in_bw, -- 满足func+'-State'规则
        avgMerge(out_bw) AS out_bw,
        avgMerge(max_bw) AS max_bw
    FROM
        pili_hera.bw_nerve_aggre 
    GROUP BY en_name, real_server_id,record_time,isp
    ) a
join isp_map
on a.isp = isp_map.isp_id;
```