package config

import "fmt"

type DBConfig struct {
	Host string `mapstructure:"DB_HOST"`
	Port string `mapstructure:"DB_PORT"`
	User string `mapstructure:"POSTGRES_USER"`
	Pass string `mapstructure:"POSTGRES_PASSWORD"`
	Name string `mapstructure:"POSTGRES_DB"`
	Path string `mapstructure:"DB_PATH"`
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

func (dbConfig *DBConfig) GetPath() string {
	return dbConfig.Path
}
