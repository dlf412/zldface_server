package config

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"gorm.io/gorm"
	"log"
	"os"
	"strings"
)

type system struct {
	Debug            bool  `yaml:"debug"`
	Addr             int   `yaml:"addr"`
	MultiPoint       bool  `yaml:"multipoint"`
	MultipartMemory  int64 `yaml:"multipartmemory"`
	MatchConcurrency int   `yaml:"matchconcurrency"`
}

type storage struct {
	RegDir string `yaml:"regdir"`
	VerDir string `yaml:"verdir"`
}

type oauth2 struct {
	SuperToken string `yaml:"superToken"`
}

type arcsoft struct {
	ExpiredAt string `yaml:"expiredAt"`
	AlarmDays int    `yaml:"alarmDays"`
}

type Cfg struct {
	Redis   redis_cfg
	System  system
	Zap     zap_cfg
	Mysql   mysql_cfg
	Storage storage
	Auth    string
	OAuth2  oauth2
	Arcsoft arcsoft
}

var Config = Cfg{}
var RedisCli *redis.Client
var Logger = ZldLog{}
var DB *gorm.DB
var RegDir, VerDir string
var Debug, MultiPoint bool

func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	for _, k := range viper.AllKeys() {
		value := viper.GetString(k)
		if strings.HasPrefix(value, "${") && strings.HasSuffix(value, "}") {
			viper.Set(k, getEnv(strings.TrimSuffix(strings.TrimPrefix(value, "${"), "}")))
		}
	}

	err := viper.Unmarshal(&Config)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}
	MultiPoint = Config.System.MultiPoint

	if MultiPoint {
		RedisCli = Config.Redis.Init()
	}
	Logger.Logger = Config.Zap.Init()
	DB = Config.Mysql.Init()

	VerDir = Config.Storage.VerDir
	RegDir = Config.Storage.RegDir
	Debug = Config.System.Debug

	if !Debug {
		gin.SetMode(gin.ReleaseMode)
	}
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
