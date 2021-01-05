package config

import (
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"log"
	"os"
	"strings"
)

type system struct {
	Debug bool `yaml:"debug"`
	Addr int `yaml:"addr"`
}

type Cfg struct {
	Redis redis_cfg
	System system
	Zap zap_cfg
}

var Config Cfg = Cfg{}
var RedisCli *redis.Client
var Logger *zap.Logger

func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	for _, k := range viper.AllKeys() {
		value := viper.GetString(k)
		//log.Println(k, value)
		if strings.HasPrefix(value, "${") && strings.HasSuffix(value, "}") {
			viper.Set(k, getEnv(strings.TrimSuffix(strings.TrimPrefix(value,"${"), "}")))
		}
	}

	err := viper.Unmarshal(&Config)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}

	log.Println(Config)

	RedisCli = Config.Redis.Init()
	Logger = Config.Zap.Init()
}

func getEnv(env string) string {
	// ENV:default
	env_default := strings.SplitN(env, ":", 2)
	res := os.Getenv(env_default[0])
	if len(res) == 0 && len(env_default) > 1 {
		res = env_default[1]
	}
	return res
}