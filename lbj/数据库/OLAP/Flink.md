# Flink

基于操作符(Operator)的连续流模型，可以做到微妙级别的延迟。而Flink基于每个事件处理，每当有新的数据输入都会立刻处理，是真正的流式计算，支持毫秒级计算。

## SQL

### hints

[spark中的join](https://blog.csdn.net/wlk_328909605/article/details/82933552)

```sql
-- 通过OPTIONS注释的方式更改表配置
select id, name from kafka_table1 /*+ OPTIONS('scan.startup.mode'='earliest-offset') */;
-- SHUFFLE_HASH 只支持等值联接条件，即选中build的小表的联接key做hashmap，映射连接。
-- 联接提示会失效，只能使用支持非等值条件联接的 nested loop join。
SELECT /*+ SHUFFLE_HASH(t1) */ * FROM t1 join t2 ON t1.id > t2.id;
```

## 聚合/窗口聚合

窗口聚合表值函数包括TUMBLE(),HOP(),CUMULATE()

```sql
-- group 支持group setting写法
-- rollup是对包含数组及其前缀子集的grouping set的缩写
-- CUBE则表示其所有子集的缩写
-- 注意，窗口聚合需要把窗口函数相关值聚合，其他再使用缩写grouping set
SELECT window_start, window_end, supplier_id, SUM(price) as price
FROM TABLE(
    TUMBLE(TABLE Bid, DESCRIPTOR(bidtime), INTERVAL '10' MINUTES))
GROUP BY window_start, window_end, ROLLUP (supplier_id);
```

窗口聚合后的结果，窗口聚合表值函数返回具有时间属性window_time，值为window_end的前0.001s

开窗函数：`OVER( [ PARTITION BY … ] [ ORDER BY … ] [RANGE ...])` 对聚合函数进行分组，排序，筛选

## join

### regular join

普通的内外联接语法符合sql标准，但由于左右表的值发生变化后，均会触发更新。如果流存储的历史数据过大，将会引起异常。

### Event Time Temporal Join

Event Time Temporal joins仅仅支持对版本表进行join，即可以通过检索还原当时的表数据，且append-only。在join时，左表可以是任意表(但需要包含watermark触发join计算)，但右表是一张版本表，flink使用`FOR SYSTEM_TIME AS OF`的sql语法实现。右边的版本表，应该存储在watermark到来前的所有版本数据，这样左表数据到来时，都能join出正确数据。历史左表数据到来时，不会触发join，而是会被缓存起来，直到新的watermark被触发后一起输出(所以历史数据不宜过多，这样会占用过多缓存空间，缓存应该是占有内存，未看见flink有描述持久化功能)。一般情况下，该join都是append-only的插入操作，如果是不当的修改删除操作，会输出retract流的带del的changelog。case如下：

```sql
SELECT 
     order_id,
     price,
     currency,
     conversion_rate,
     order_time
FROM orders
LEFT JOIN currency_rates FOR SYSTEM_TIME AS OF orders.order_time -- 选择时间点的ddl写法，此处为选择事件时间，但注意不能选择当前的处理时间
ON orders.currency = currency_rates.currency;
```

注意：build side侧数据改变不会影响之前的数据

### Processing Time Temporal Join

通常用来联接无法物化为flink动态表的维表如mysql,hbase，维表作为右表构建表进行连接。由于兼容性因素，使用Temporal Table Function(udtf实现，似乎只能自定义且有java实现)来实现定位最新处理时间从而连接维表，而不能使用ddl语法(至少1.13版本前可以)，只支持内联接和左外连接。case如下：

```sql
SELECT
  o_amount, r_rate
FROM
  Orders,
  LATERAL TABLE (Rates(o_proctime)) -- temporal table functions定位最新时间
WHERE
  r_currency = o_currency
```

注意：1.build side侧数据改变不会影响之前的数据 2.并不会存储维表的状态

### Lookup Join

为mysql类的数据库使用，使用条件:1.左表具有processing时间属性。2.数据源的connector具有look up功能。case如下：

```sql
-- Customers is backed by the JDBC connector and can be used for lookup joins
-- 临时表结构，可以用来现用现查现缓存，而不是flink一直维持缓存状态
CREATE TEMPORARY TABLE Customers (
  id INT,
  name STRING,
  country STRING,
  zip STRING
) WITH (
  'connector' = 'jdbc',
  'url' = 'jdbc:mysql://mysqlhost:3306/customerdb',
  'table-name' = 'customers'
);
-- enrich each order with customer information
SELECT o.order_id, o.total, c.country, c.zip
FROM Orders AS o
  JOIN Customers FOR SYSTEM_TIME AS OF o.proc_time AS c
    ON o.customer_id = c.id;
```

## UDF

### ScalarFunction 标量函数

标量函数的作用是将多个标量转化为一个新值

### TableFunctions 表函数

用于表连接后的复杂转换，返回值

### AggregationFunctions 聚合函数

复杂聚合计算的函数

### 窗口

滚动窗口(Tumble Windows)，滑动窗口(Hop Window)，会话窗口。

flink的窗口是通过复制实现的，如果同时存在100个窗口的话，则事件存在在所有窗口中。如多个滑动窗口，watermark大大延迟的滚动窗口。
时间窗口将会根据实际时间对齐，而且不是开启窗口时间，相当于整点对齐。
时间窗口接时间窗口，时间戳将会自动使用上一个窗口计算的结束时间。
会话窗口的实现，是通过小窗口结果的合并，如果延迟事件能弥补窗口之间的间隙。

### 时间属性

`事件时间（Event Time）----> 提取时间（Ingestion Time）----> 处理时间（Processing Time）`

事件时间：服务记录采集的该事件发生的时间，如点击事件发生时的产生时间。
提取时间：数据进入flink框架的时间。
处理时间：数据流入具体某个算子时候对应的系统时间，所有基于时间的操作都是使用各物理机上的系统时间。

## 精确一次计算

1.at most once 2.at least once 3.at least once

Flink可通过回退和重新发送source的数据流从故障中恢复，flink支持如上三种计算输出。当理想情况被描述为精确一次时，必须满足两个条件：可重放的source，事务性/可幂等的sink。一般采用2提高性能。

## 时态表

时态表（Temporal Table）是一张随时间变化的表，包含一个或多个版本的表快照集合，在flink中被称为动态表。如果时态表中的记录可以追踪和并访问它的历史版本，这种表我们称之为版本表，如来自数据库的changelog可以定义成版本表(包含增删改信息，增量表包含历史信息)。如果时态表中的记录仅仅可以追踪并和它的最新版本，这种表我们称之为普通表，来自数据库或HBase的表可以定义成普通表(全量表只包含当前信息)。

### flink中的版本表与普通表

在flink中，版本表和普通表的区别在于定义了主键约束和事件时间属性。如下：

```sql
PRIMARY KEY(product_id) NOT ENFORCED,      -- (1) 定义主键约束
WATERMARK FOR update_time AS update_time   -- (2) 通过 watermark 定义事件时间  

-- RatesHistory是一个append-only的只带时间属性的表
-- 使用特殊的去重查询获取主键，构建版本视图，这样flink就可以通过kafka输出动态表的changelog。
CREATE VIEW versioned_rates AS              
SELECT currency, rate, currency_time            -- (1) `currency_time` 保留了事件时间
  FROM (
      SELECT *,
      ROW_NUMBER() OVER (PARTITION BY currency  -- (2) `currency` 是去重 query 的 unique key，可以作为主键
         ORDER BY currency_time DESC) AS rowNum 
      FROM RatesHistory )
WHERE rowNum = 1; 
```

### temporal table functions

temporal table functions提供可以访问某个时间点表值的方式，但值得注意的是只有定义在append-only流之上的时态表才能使用(右表)。使用方式如下：

```sql
SELECT
  SUM(amount * rate) AS amount
FROM
  orders,
  LATERAL TABLE (rates(order_time)) -- 表函数输入rates表以及其时间属性
WHERE
  rates.currency = orders.currency
```

### 表到流的转换

完成物化视图的即时维护来实现动态表。一般的物化视图的即时维护，只响应基表的插入操作，更新和删除操作维护起来比较复杂。在使用时间属性的连续查询中可以做到只有插入操作，而不使用时间属性的操作将会导致更新，删除操作。将连续查询结果的动态表转换为流或将其写入外部系统时：

1. append-only流：只包含insert操作。(按时间属性聚合的插入表，只做插入的物化视图)
2. retract流：通过add和retract(删除)操作完成dml，更新操作由先删再增逻辑实现。(不按时间属性聚合，包含更新、删除触发的物化视图)
3. upsert流：通过唯一索引标识更新和插入绑定，delete操作分离。(不按时间属性聚合，包含更新、删除触发的物化视图)

## 不确定性

### 函数的不确定性

非确定性函数：在运行时，对每一条记录进行计算，如：UUID，RAND，RAND_INTEGER，CURRENT_DATABASE...
动态函数：在生成执行计划期间被确定，该计划的记录值均相同，如：CURRENT_DATE，CURRENT_TIME，CURRENT_TIMESTAMP，NOW，LOCALTIME...

### 流上的不确定性

source连接器回溯读取时的不确定性：如在kafka中，在相同时间点位读取数据可能会因为过期而失效等问题。(可以持久化存储，然后重发kafka)
基于处理时间的不确定性：如采用处理时间来使用依赖时间属性的窗口聚合，还原历史时join的维表发生变动等信息。(使用事件时间，维表数据保存历史存档，或者使用明细表数据进行还原)
基于TTL淘汰内部状态数据的不确定性：分组聚合(非窗口聚合)等维护内部状态的数据持续膨胀，需要开启ttl过期来淘汰数据，这样结果就不确定了。