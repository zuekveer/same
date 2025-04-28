package config

import (
	"bytes"
	"embed"
	"fmt"

	"app/internal/logger"

	"github.com/spf13/viper"
)

//go:embed config.yaml
var embeddedConfig embed.FS

type Config struct {
	DB      DBConfig
	App     AppConfig
	Metrics MetricsConfig
}

type DBConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	Name     string
}

type AppConfig struct {
	Port string
}

type MetricsConfig struct {
	Port string
}

func LoadConfig() (Config, error) {
	v := viper.New()

	content, err := embeddedConfig.ReadFile("config.yaml")
	if err != nil {
		logger.Logger.Warn("Failed to read embedded config.yaml: %v", err)
	} else {
		v.SetConfigType("yaml")
		if err := v.ReadConfig(bytes.NewBuffer(content)); err != nil {
			logger.Logger.Warn("Failed to load embedded config.yaml: %v", err)
		}
	}

	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		logger.Logger.Warn("Unable to decode config into struct: %v", err)
	}

	return cfg, nil
}

func (db DBConfig) ConnString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", db.User, db.Password, db.Host, db.Port, db.Name)
}
