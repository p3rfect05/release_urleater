package config

import "fmt"

type DBConfig struct {
	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDatabase string
	PostgresParams   string
}
type Config struct {
	DB DBConfig
}

func (c *Config) PostgresURL() string {
	pgURL := fmt.Sprintf(
		"postgres://%v:%v@%v:%v/%v",
		c.DB.PostgresUser,
		c.DB.PostgresPassword,
		c.DB.PostgresHost,
		c.DB.PostgresPort,
		c.DB.PostgresDatabase,
	)

	if c.DB.PostgresParams != "" {
		pgURL = fmt.Sprintf("%v?%v", pgURL, c.DB.PostgresParams)
	}

	return pgURL
}
