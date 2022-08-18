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

## 规模较小时

```
//使用dump的方式进行备份
clickhouse-client -d database --query="SELECT * FROM [db.]tablename format CSV" > export_tablename.csv
//进行插入工作
cat export_tablename.csv | clickhouse-client --query="INSERT INTO [db.]tablename FORMAT CSV";
```