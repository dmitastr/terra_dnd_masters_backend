package config

import "fmt"

type DBConfig struct {
	Host string `mapstructure:"DB_HOST"`
	Port string `mapstructure:"DB_PORT"`
	User string `mapstructure:"DB_USER"`
	Pass string `mapstructure:"DB_PASSWORD"`
	Name string `mapstructure:"DB_NAME"`
}

func (dbConfig *DBConfig) GetConnString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbConfig.User,
		dbConfig.Pass,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.Name,
	)
}
