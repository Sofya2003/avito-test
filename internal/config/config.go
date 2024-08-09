package config

type Config struct {
	PostgresURL string
	APIPort     int
}

func NewConfig() (config *Config) {
	return &Config{
		PostgresURL: "postgres://postgres:postgres@postgres:5432/postgres?sslmode=disable",
		APIPort:     8080,
	}
}
