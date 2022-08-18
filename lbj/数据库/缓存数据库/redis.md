# redis

## [redis下载](https://redis.io/docs/getting-started/installation/install-redis-on-mac-os/)

## 常用指令

Hgetall:返回该key的所有的字段和值，所以返回长度为哈希表大小的两倍

```
redis> HSET myhash field1 "Hello"
(integer) 1
redis> HSET myhash field2 "World"
(integer) 1
redis> HGETALL myhash
1) "field1"
2) "Hello"
3) "field2"
4) "World"
```