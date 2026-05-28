package config

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/spf13/viper"
)

type ConfigProvider interface {
	GetAddress() string
	GetDBConfig() *DBConfig
}

type Config struct {
	AppHost  string `mapstructure:"HOST"`
	AppPort  string `mapstructure:"PORT"`
	DbConfig *DBConfig
}

func NewConfig() (*Config, error) {
	viper.AutomaticEnv()

	viper.BindEnv("PORT")
	viper.BindEnv("HOST")
	viper.BindEnv("DB_PORT")
	viper.BindEnv("DB_HOST")
	viper.BindEnv("DB_USER")
	viper.BindEnv("DB_PASSWORD")
	viper.BindEnv("DB_NAME")

	var config Config
	var dbConfig DBConfig

	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	if err := viper.ReadInConfig(); err != nil {
		if !errors.As(err, &viper.ConfigFileNotFoundError{}) {
			return nil, fmt.Errorf("error reading config file, %s", err)
		}
	}
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode config, %v", err)
	}
	if err := viper.Unmarshal(&dbConfig); err != nil {
		return nil, fmt.Errorf("unable to decode dbConfig, %v", err)
	}

	if config.AppHost == "" {
		config.AppHost = "localhost"
	}
	if config.AppPort == "" {
		config.AppPort = "8080"
	}

	config.DbConfig = &dbConfig

	log.Printf("%#v\n", config)
	log.Printf("%#v\n", dbConfig)
	return &config, nil
}

func (c *Config) GetAddress() string {
	return fmt.Sprintf("%s:%s", c.AppHost, c.AppPort)
}

func (c *Config) GetDBConfig() *DBConfig {
	return c.DbConfig
}

// getEnv get key environment variable if exists, otherwise return defaultValue
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}
	return value
}
