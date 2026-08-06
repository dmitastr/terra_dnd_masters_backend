package config

import (
	"errors"
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type ConfigProvider interface {
	GetAddress() string
	GetDBConfig() *DBConfig
	GetAPIKey() string
}

type Config struct {
	AppHost      string `mapstructure:"HOST"`
	AppPort      string `mapstructure:"PORT"`
	ClientAPIKey string `mapstructure:"CLIENT_API_KEY"`
	DbConfig     *DBConfig
}

func NewConfig() (*Config, error) {
	viper.AutomaticEnv()

	_ = viper.BindEnv("PORT")
	_ = viper.BindEnv("HOST")
	_ = viper.BindEnv("DB_PORT")
	_ = viper.BindEnv("DB_HOST")
	_ = viper.BindEnv("POSTGRES_USER")
	_ = viper.BindEnv("POSTGRES_PASSWORD")
	_ = viper.BindEnv("POSTGRES_DB")
	_ = viper.BindEnv("CLIENT_API_KEY")

	var config Config
	var dbConfig DBConfig

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
		config.AppHost = "0.0.0.0"
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

func (c *Config) GetAPIKey() string {
	return c.ClientAPIKey
}
