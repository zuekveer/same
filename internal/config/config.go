package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	DB  DBConfig
	App AppConfig
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

func LoadConfig() Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")        // current dir
	viper.AddConfigPath("./config") // fallback dir

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("Unable to decode config into struct: %v", err)
	}

	return cfg
}

func (db DBConfig) ConnString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", db.User, db.Password, db.Host, db.Port, db.Name)
}
