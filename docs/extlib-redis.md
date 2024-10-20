# Module - "redis"

```golang
redis := import ("redis")
```

## Functions

- `dial(url)=>RedisClient`: 连接到redis库，其中`url`格式：`redis://user:password@localhost:6789/3?dial_timeout=3&db=1&read_timeout=6s&max_retries=2`

## RedisClient

- `set(key,value,[ttl])=>StatusCmd`: 设置redis值,TTL值默认1小时。如果要设置更长的时间，参考：
  `set(key,value,5*times.hour)`,如果设置永不过期：  `set(key,value,-1)`
- `setnx(key,value)=>StatudCmd`: 参考set的设置，通常用于事务性的操作
- `get(key)=>StringCmd`: 获取redis中的值