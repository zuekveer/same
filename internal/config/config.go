package config

import (
	"fmt"

	"app/internal/logger"

	"github.com/spf13/viper"
)

type Config struct {
	DB  DBConfig  `mapstructure:"db"`
	App AppConfig `mapstructure:"app"`
}

type DBConfig struct {
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Name     string `mapstructure:"name"`
}

type AppConfig struct {
	Port string `mapstructure:"port"`
}

func LoadConfig() Config {
	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("./internal/config")
	v.AddConfigPath("/app/internal/config")

	if err := v.ReadInConfig(); err != nil {
		logger.Logger.Info("Warning: could not read config.yaml: %v", err)
	}

	v.SetConfigFile(".env")
	if err := v.MergeInConfig(); err != nil {
		logger.Logger.Info("No .env file found or failed to merge .env")
	}

	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		logger.Logger.Warn("Unable to decode config into struct: %v", err)
	}

	return cfg
}

func (db DBConfig) ConnString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", db.User, db.Password, db.Host, db.Port, db.Name)
}
