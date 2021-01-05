package config

import "github.com/go-redis/redis/v8"

type redis_cfg struct {
	Url string  `yaml:"url"`
	DB  string  `yaml:"db"`
}

func (r redis_cfg) Init() *redis.Client {
	opt, err := redis.ParseURL(r.Url + r.DB)
	if err != nil {
		panic(err)
	}
	rdb := redis.NewClient(opt)
	return rdb
}




