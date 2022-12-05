# RDBMS

## Table Valued Function

表值函数（Table Valued Function），从字面意义上可以看出，其返回值是Table类型，是一个数据集，可以简单理解为带有参数的视图。即将视图查询的sql，通过将某些查询字段参数化，由不同入参组合出不同的查询语句。可以将不同表的查询结果，汇总到一张虚拟的表值数据集中，如a,b两张不同的表分别查询相同字段的数据，一起返回。

## grouping set

```sql
-- 将不同分组条件的查询结果union all，一起输出
SELECT
    warehouse,
    product, 
    SUM (quantity) qty
FROM
    inventory
GROUP BY
    GROUPING SETS(
        (warehouse,product),
        (warehouse),
        (product),
        ()
    );
```