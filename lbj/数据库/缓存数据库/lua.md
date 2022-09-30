# lua

case

```lua
-- KEYS[1]表示hash key,ARGV[1],ARGV[2]表示校验field，ARGV[3]表示想校验的值
local v = redis.call('HGET',KEYS[1],ARGV[1])
if v == nil then
    return nil
elseif v ~= ARGV[3] then
    redis.call('DEL',KEYS[1])
    return nil
else 
    return redis.call('HGET',KEYS[1],ARGV[2])
end
```

注意:false == nil == 'nil'