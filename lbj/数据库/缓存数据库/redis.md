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

## 批量指令

特点：具备原子性

SET->MSET,
GET->MGET,
LSET->LPUSH,RPUSH,
LINDEX->LRANGE,
HSET->HMSET,
HGET->HMGET

DEL key1 key2 key3  //不支持通配符，如果该key不存在则忽略，返回删除个数

## pipeline

为了减少RTT(round trip time - 往返时间)可以使用pipeline，但是pipeline存在如下问题：

1. pipeline不具备原子性，其中一个命令执行异常，依然会继续执行后续指令。
2. 不具备隔离性，即执行的过程可被拆分，且可被其他客户端指令穿插执行。redis服务端的指令缓冲队列如果同时有多个客户端发送指令，并行入队时并不能保证被pipeline指令独占。
3. 与lua脚本相比不能添加指令之间的依赖逻辑关系
4. socket-outpot缓冲区一般大小为8k，pipeline的指令数受到限制
5. cluster集群不支持pipeline操作 -- 基本可以判断该命令死刑了

## 事务与Bxx指令

redis的事务主要是为了满足隔离性，即这些命令在执行时，不允许其他客户端插队，因此在某些简单Bxx指令就是对应事务状态的xx指令。事务的开启与结束通过MULTI/EXEC指令实现，在执行MULTI命令时标记事务开始，期间的客户端发送的多条命令只会缓存，不会立刻执行，在调用EXEC指令后一次性执行。如果调用了discard指令，则会清空事务队列，并推出事务状态。

watch指令执行在事务开启之前，在事务execute时，会通过cas操作检查watch相关值是否发生了变化，如果发生了变化则取消事务的执行。其中watch监控的变更期主要为，使用watch指令到execute之间，在execute执行完毕无论成功失败，watch都会取消。由此可以看出，redis内部的cas操作反应到程序代码就是，业务代码也需要设置乐观锁多次重试。

事务处理的情况：

1. 出现指令语法错误，则不执行该事务
2. 出现类型相关不匹配错误，忽略该条指令继续执行
3. 出现watch值变化，取消执行该事务

redis集群不支持事务操作，因为事务可能会跨槽，影响很多台机器。

## lua脚本

```
EVAL luascript numkeys key [key ...] arg [arg ...]
ex:
EVAL "return {KEYS[1],KEYS[2],ARGV[1],ARGV[2]}" 2 key1 key2 first second
EVAL "return redis.call('GET',KEYS[1])" 1 hello

SCRIPT LOAD script // 缓存脚本，但不立即执行
SCRIPT FLUSH // 清空缓存脚本
EVALSHA sha1 numkeys key [key ...] arg [arg ...] // 使用缓存的sha1值执行脚本
SCRIPT EXISTS script [script ...] // 检查脚本是否已缓存
SCRIPT KILL // 杀死当前正在运行的lua脚本
```

### 集群问题

按理说可以直接将kv值写入脚本，但Redis官方文档指出这种是不建议的，目的是在命令执行前会对命令进行分析，以确保Redis Cluster可以将命令转发到适当的集群节点。特别是现在对跨槽操作进行了限制，会对k进行分析，判断是否跨槽。

### 原子问题

Lua脚本在Redis中是以原子方式执行的，在Redis服务器执行EVAL命令时，在命令执行完毕并向调用者返回结果之前，只会执行当前命令指定的Lua脚本包含的所有逻辑，其它客户端发送的命令将被阻塞，直到EVAL命令执行完毕为止。因此LUA脚本不宜编写一些过于复杂了逻辑，必须尽量保证Lua脚本的效率，否则会影响其它客户端。

### 要求

* 务必对Lua脚本进行全面测试以保证其逻辑的健壮性，当Lua脚本遇到异常时，已经执行过的逻辑是不会回滚的。
* 尽量不使用Lua提供的具有随机性的函数，参见相关官方文档。
* 在Lua脚本中不要编写function函数,整个脚本作为一个函数的函数体。
* 在脚本编写中声明的变量全部使用local关键字。
* 在集群中使用Lua脚本要确保逻辑中所有的key分到相同机器，也就是同一个插槽(slot)中，可采用Redis Hash Tag技术。
* 再次重申Lua脚本一定不要包含过于耗时、过于复杂的逻辑。

## 集群

### 要求

1. 在没有代理，没有异步同步，value合并操作的条件下能线性拓展到1000个节点
2. 可以接受的写安全
3. 可用性要求集群的大部分主都能可达，不能可达的主节点至少有一个副本可达。并且当主节点不再拥有副本时，其他拥有副本的主节点将转移一个副本给该节点。

## 数据结构

[zset的底层结构](https://zhuanlan.zhihu.com/p/193141635)
