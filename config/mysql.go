package config

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)

type mysql_cfg struct {
	Host         string `mapstructure:"host" json:"host" yaml:"host"`
	Port         string `mapstructure:"port" json:"port" yaml:"port"`
	DB           string `mapstructure:"db" json:"db" yaml:"db"`
	Config       string `mapstructure:"config" json:"config" yaml:"config"`
	User         string `mapstructure:"user" json:"user" yaml:"user"`
	Password     string `mapstructure:"password" json:"password" yaml:"password"`
	MaxIdleConns int    `mapstructure:"max-idle-conns" json:"maxIdleConns" yaml:"max-idle-conns"`
	MaxOpenConns int    `mapstructure:"max-open-conns" json:"maxOpenConns" yaml:"max-open-conns"`
	LogMode      bool   `mapstructure:"log-mode" json:"logMode" yaml:"log-mode"`
	LogZap       string `mapstructure:"log-zap" json:"logZap" yaml:"log-zap"`
}

func Init(m mysql_cfg) *gorm.DB {
	dsn := m.User + ":" + m.Password + "@tcp(" + m.Host + ":" + m.Port + ")/" + m.DB + "?" + m.Config
	mysqlConfig := mysql.Config{
		DSN:                       dsn,   // DSN data source name
		DefaultStringSize:         80,    // string 类型字段的默认长度
		DisableDatetimePrecision:  true,  // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		DontSupportRenameIndex:    true,  // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		DontSupportRenameColumn:   true,  // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
		SkipInitializeWithVersion: false, // 根据版本自动配置
	}
	if db, err := gorm.Open(mysql.New(mysqlConfig), gormConfig(m.LogMode)); err != nil {
		panic(err)
	} else {
		sqlDB, _ := db.DB()
		sqlDB.SetMaxIdleConns(m.MaxIdleConns)
		sqlDB.SetMaxOpenConns(m.MaxOpenConns)
		return db
	}
}

func gormConfig(mod bool) *gorm.Config {
	var config = &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true}
	var log_conf = logger.Config{
		SlowThreshold: 200 * time.Millisecond,
		LogLevel:      logger.Warn,
		Colorful:      true,
	}
	switch Config.Mysql.LogZap {
	case "silent", "Silent":
		config.Logger = logger.New(Logger, log_conf).LogMode(logger.Silent)
	case "error", "Error":
		config.Logger = logger.New(Logger, log_conf).LogMode(logger.Error)
	case "warn", "Warn":
		config.Logger = logger.New(Logger, log_conf).LogMode(logger.Warn)
	case "info", "Info":
		config.Logger = logger.New(Logger, log_conf).LogMode(logger.Info)
	case "zap", "Zap":
		config.Logger = logger.New(Logger, log_conf).LogMode(logger.Info)
	default:
		if mod {
			config.Logger = logger.Default.LogMode(logger.Info)
			break
		}
		config.Logger = logger.Default.LogMode(logger.Silent)
	}
	return config
}
