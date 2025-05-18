package config

import (
	"bytes"
	"embed"
	"fmt"
	"log/slog"
	"time"

	"github.com/spf13/viper"
)

//go:embed config.yaml
var embeddedConfig embed.FS

type Config struct {
	App     AppConfig     `mapstructure:"app"`
	Metrics MetricsConfig `mapstructure:"metrics"`
	DB      DBConfig      `mapstructure:"db"`
	Cache   CacheConfig   `mapstructure:"cache"`
	Tracing TracingConfig `mapstructure:"tracing"`
	Logger  LoggerConfig  `mapstructure:"logger"`
}

type DBConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	Name     string
}

type LoggerConfig struct {
	Level string `mapstructure:"level"`
}

// without mapstructure tag configs doesn't work in app.go

type CacheConfig struct {
	ExpirationMinutes time.Duration `mapstructure:"expiration_minutes"`
	CleanupMinutes    time.Duration `mapstructure:"cleanup_minutes"`
}

type AppConfig struct {
	Port string
}

type MetricsConfig struct {
	Port string
}

type TracingConfig struct {
	JaegerEndpoint string `mapstructure:"jaeger_endpoint"`
}

func LoadConfig() (Config, error) {
	v := viper.New()

	content, err := embeddedConfig.ReadFile("config.yaml")
	if err != nil {
		slog.Warn("Failed to read embedded config.yaml: %v", "error", err)
	} else {
		v.SetConfigType("yaml")
		if err := v.ReadConfig(bytes.NewBuffer(content)); err != nil {
			slog.Warn("Failed to load embedded config.yaml: %v", "error", err)
		}
	}

	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		slog.Warn("Unable to decode config into struct: %v", "error", err)
	}

	return cfg, nil
}

func (db DBConfig) ConnString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", db.User, db.Password, db.Host, db.Port, db.Name)
}
