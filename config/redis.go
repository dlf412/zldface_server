package config

import "github.com/go-redis/redis/v8"
import "context"

var Rctx = context.Background()

type redis_cfg struct {
	Url string `yaml:"url"`
	DB  string `yaml:"db"`
}

func (r redis_cfg) Init() *redis.Client {
	opt, err := redis.ParseURL(r.Url + r.DB)
	if err != nil {
		panic(err)
	}
	rdb := redis.NewClient(opt)
	{
		if err := rdb.Ping(Rctx).Err(); err != nil {
			panic(err)
		}
	}
	return rdb
}
