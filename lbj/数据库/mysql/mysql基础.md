# 基础

## 配置

### mac配置

查看配置文件路径依赖：`mysqld --help --verbose | more`
一般配置依赖路径优先级：/etc/my.cnf <- /etc/mysql/my.cnf <- /usr/local/etc/my.cnf <- ~/.my.cnf

### 关闭ONLY\_FULL\_GROUP\_BY

1. 查看当前sql_mode:`select @@sql_mode`
2. 复制当前值且去掉 **ONLY_FULL_GROUP_BY**
3. `set sql_mode = '复制值'`;(这个似乎是改变当前库的，亲测有效。似乎新建的无效，新建的需要需要`set @@global.sql_mode`----每次重连失效[补充:mysql的变量分为local,session,global,配置文件四个级别。])

## mysql变量

全局变量：设置全局变量需要super权限(应该是指数据库的权限)。

```sql
SET GLOBAL sort_buffer_size=value;
SET @@global.sort_buffer_size=value;

SELECT @@global.sort_buffer_size;
SHOW GLOBAL VARIABLES like 'sort_buffer_size';
```

会话变量：等同与local变量，不需要权限验证

```sql
SET SESSION sort_buffer_size=value;
SET @@session.sort_buffer_size=value;
SET sort_buffer_size=value;

SELECT @@sort_buffer_size;
SELECT @@session.sort_buffer_size;
SHOW SESSION VARIABLES like 'sort_buffer_size';
```

临时变量：预处理，这样操作应该是把结果存在mysql中，而不必将它们存储在客户端的临时变量中

```sql
SELECT @min_price:=MIN(price),@max_price:=MAX(price) FROM shop;
SELECT * FROM shop WHERE price=@min_price OR price=@max_price;
```

## 启动

### [官网下载](dev.mysql.com/downloads/mysql/5.7.html)

1. 可以通过mac的系统偏好设置，选择mysql启动服务
2. 终端输入命令`sudo /usr/local/mysql/support-files/mysql.server start` 启动

其他：
(**注意第一次操作时会失效，可能是安装时启动没有录入pid，在mac服务里面关闭mysql服务后即可正常使用指令。使用mac启动服务就是没有pid**)
停止mysql:`sudo /usr/local/mysql/support-files/mysql.server stop`  
重启mysql:`sudo /usr/local/mysql/support-files/mysql.server restart`

### brew下载

1. brew services start mysql //启动mysql
2. brew services stop mysql //停止mysql
3. brew services restart mysql //重新启动mysql

## [初始化修改密码](https://www.jianshu.com/p/1def4f9c4ecf)

## 连接注意

1. left/right join以相应边的表为基础进行连接，这样不满足条件的右边表便会以null形式展现出来。
2. (inner) join mysql会根据相关统计信息和策略选择驱动表，然后再连接保留两边都有的数据
3. 外连接包括左外，右外和全外连接，全连接mysql不支持，一般使用左外连接和右外连接union取并集获得

### 事务

```sql
START TRANSACTION or BEGIN start a new transaction.

COMMIT commits the current transaction, making its changes permanent.

ROLLBACK rolls back the current transaction, canceling its changes.

SET autocommit disables or enables the default autocommit mode for the current session.

READ UNCOMMITTED、READ COMMITTED、REPEATABLE READ、SERIALIZABLE            //事务支持的隔离级别
SET TRANSACTION ISOLATION LEVEL SERIALIZABLE;   //设置事务的隔离级别，但是不允许此时有事务正在执行。
SET GLOBAL TRANSACTION SERIALIZABLE;            //注意，似乎只对下次事务一次有效(上面的语句一样的)
```

### 索引修改

索引相关指令：

1. 新增索引:`ALTER TABLE tb_name ADD [INDEX|UNIQUE|FULLTEXT index_name] | [PRIMARY KEY] (username(length)...)` //都需要指定长度。
2. 删除索引:`ALTER TABLE tb_name DROP [INDEX i | PRIMARY KEY]`
3. 修改索引依赖列为非空:`alter table tb_name modify column_name type(length) not null`
4. 查看索引`show index from tb_name [from db_name]`
5. 查看表的存储引擎`show create table tb_name`;

>
字段解析：  
Non\_unique:表示是否为唯一索引，是为0否为1。  
Null:表示是否含有Null，是为Yes否为NO。  
Key\_name:表示索引名。  
Column\_name,Seq\_in_index:分别表示列名，以及该列在索引中的位置如果为单列则改值为1，为组合索引则为索引的顺序。  
Collation:表示排序顺序显示"A"为升序，为NULL则表示无分裂。  
Cardinality:表示索引中唯一值数量的估算数量，基数越大，当进行联合时使用该索引的机会越大(???)。  
Sub\_part:表示列中被编入索引的字符的数量，如整列被编入则为NULL。  
Packed:指示关键字如何被压缩，若没被压缩则为NULL。  

### 数据导出

## 数据库，表操作

```
database db_name;  -- 创建数据库
show databases;           -- 显示所有的数据库
drop database db_name;    -- 删除数据库
use db_name;              -- 选择数据库
create table tb_name (字段名 varchar(20), 字段名 char(1));   -- 创建数据表模板,注意使用default时，字符串这边似乎有确定的要求'',必须使用单引号。
show tables;              -- 显示数据表
desc tb_name；            -- 显示表结构,等同show columns tb_name;
show all columns tb_name;         --  显示详细表信息，包含comment等
drop table tb_name；      -- 删除表
ALTER TABLE <表名> [修改选项]  -- 修改表
[修改选项]：
[ ADD COLUMN <列名> <类型>
| CHANGE COLUMN <旧列名> <新列名> <新列类型>
| ALTER COLUMN <列名> { SET DEFAULT <默认值> | DROP DEFAULT }
| MODIFY COLUMN <列名> <类型>
| DROP COLUMN <列名>
| RENAME TO <新表名>
| CHARACTER SET <字符集名>
| COLLATE <校对规则名> ]

alter table <表名> modify <列名> int auto_increment;	-- 设置为自增
```

### 基本语句

```sql
insert into tb_name (column1,column2,column3,...)[可省略]
 values (value1,value2,value3,...);  --插入
```

## 模糊搜索匹配

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

 ## [mysql函数](https://www.w3schools.cn/mysql/func_mysql_coalesce.asp)

* COALESCE()函数，返回函数内第一个非空值