package config

import (
	"activecounter/pkg/log"

	"github.com/spf13/viper"
)

var (
	Server *serverSetting
	Db     *dbSetting
	Redis  *redisSetting
)

type config struct {
	Server *serverSetting
	Db     *dbSetting
	Redis  *redisSetting
}

type serverSetting struct {
	Port    uint32 `json:"port,omitempty"`
	RunMode string `json:"run_mode,omitempty"`
}

type dbSetting struct {
}

type redisSetting struct {
}

func InitConfig() error {
	viper.SetConfigName("config")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %v", err)
		return err
	}

	var c config
	if err := viper.Unmarshal(&c); err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
		return err
	}

	Server = c.Server
	Db = c.Db
	Redis = c.Redis

	return nil
}
