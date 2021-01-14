package cache

/**
***基于单节点redis 分布式锁
**/

import (
	"crypto/rand"
	"github.com/go-redis/redis/v8"
	"runtime"
	"time"
	"zldface_server/config"
)

type RedisLock struct {
	lockKey string
	value   []byte
	timeout time.Duration
	loop    time.Duration
}

//保证原子性（redis是单线程），避免del删除了，其他client获得的lock
var delScript = redis.NewScript(`
if redis.call("get", KEYS[1]) == ARGV[1] then
	return redis.call("del", KEYS[1])
else
	return 0
end`)

var keepScript = redis.NewScript(`
if redis.call("get", KEYS[1]) == ARGV[1] then
	return redis.call("set", KEYS[1], ARGV[1], KEYS[2], ARGV[2])
else
	return ""
end`)

func (this *RedisLock) Lock() {

	{ //随机数
		b := make([]byte, 16)
		_, err := rand.Read(b)
		if err != nil {
			return
		}
		this.value = b
	}
	for {
		ok, err := config.RedisCli.SetNX(config.Rctx, this.lockKey, this.value, this.timeout).Result()
		if err != nil {
			return
		}
		if !ok {
			if this.loop > 0 {
				time.Sleep(time.Millisecond)
			}
			runtime.Gosched()
		} else {
			break
		}
	}
}

func (this *RedisLock) Unlock() {
	delScript.Eval(config.Rctx, config.RedisCli, []string{this.lockKey}, this.value)
}

func (this *RedisLock) Keep() bool {
	r, err := keepScript.Eval(config.Rctx, config.RedisCli, []string{this.lockKey, "px"}, this.value, int64(this.timeout/time.Millisecond)).Result()
	return err == nil && r.(string) == "OK"
}
