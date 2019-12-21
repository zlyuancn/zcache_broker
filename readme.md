# 高性能的缓存经理

---

# 获得
`go get -u github.com/zlyuancn/zcache_broker`

# 解决缓存击穿

+ 当有多个进程同时获取一个key时, 只有一个进程会真的去缓存db读取或从db加载并返回结果, 其他的进程会等待该进程结束直接收到结果.
+ 实现方式请阅读 [github.com/zlyuancn/zsingleflight](https://github.com/zlyuancn/zsingleflight)

# 解决缓存雪崩

+ 可以为命名空间设置随机的TTL, 并且可以设置该空间下的所有key每次被获取时自动刷新TTL, 可以有效减小缓存雪崩的风险

# 缓存穿透如何解决?

+ 我认为缓存穿透不是由缓存经理来解决, 而是由接收用户请求的程序来过滤这些一定不会存在的key.
+ 或许以后我们会为缓存经理添加本地内存缓存功能, 它将拥有较短的TTL, 它在同时出现大量请求同一个key(不管它是否存在)的情况下会非常有效. 但是无论如何想要减少缓存穿透的风险你都应该在用户请求key的时候判断它是否可能不存在.

# 数据db
##### 数据db用于永久性存放数据, 本模块不关心用户如何加载数据, 所以支持任何数据库

# 缓存db
##### 缓存db用于临时存放数据, 提高访问速度, 目前支持以下缓存db
+ redis
+ 任何实现 zcache_broker.CacheDB 的对象

# 连接器
##### 连接器用于提供给用户访问缓存服务的方式, 目前实现以下连接方式
+ grpc
+ rpcx


# 以下是性能测试数据
##### 10000个key, 每个key512字节随机数据, 请求key顺序随机

+grpc
```
2.50GHz * 16
goos: linux
goarch: amd64
pkg: github.com/zlyuancn/zcache_broker/transport/wgrpc/test
Benchmark_A-10          	   30000	     41881 ns/op
Benchmark_A-20          	   50000	     28340 ns/op
Benchmark_A-50          	   50000	     22747 ns/op
Benchmark_A-100         	   50000	     24701 ns/op
Benchmark_A-500         	  100000	     22383 ns/op
Benchmark_A-1000        	   50000	     28987 ns/op
Benchmark_A-5000        	   20000	     66445 ns/op
Benchmark_A-10000       	   10000	    116241 ns/op
```

+rpcx
```
2.50GHz * 16
goos: linux
goarch: amd64
pkg: github.com/zlyuancn/zcache_broker/transport/wrpcx/test
Benchmark_A-10          	   50000	     25867 ns/op
Benchmark_A-20          	  100000	     16329 ns/op
Benchmark_A-50          	  100000	     14274 ns/op
Benchmark_A-100         	  100000	     14174 ns/op
Benchmark_A-500         	  100000	     13985 ns/op
Benchmark_A-1000        	  100000	     13205 ns/op
Benchmark_A-5000        	  200000	      9668 ns/op
Benchmark_A-10000       	  200000	      7338 ns/op
```
