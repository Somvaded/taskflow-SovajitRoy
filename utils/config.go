package utils

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	Name     string `env:"NAME" envDefault:"taskflow"`
	Env      string `env:"ENV" envDefault:"development"`
	LogLevel string `env:"LOG_LEVEL" envDefault:"DEBUG"`

	// HTTP Server
	ServerPort          string        `env:"SERVER_PORT" envDefault:"8080"`
	ServerReadTimeout   time.Duration `env:"SERVER_READ_TIMEOUT" envDefault:"60s"`
	ServerWriteTimeout  time.Duration `env:"SERVER_WRITE_TIMEOUT" envDefault:"60s"`
	ServerIdleTimeout   time.Duration `env:"SERVER_IDLE_TIMEOUT" envDefault:"120s"`
	ServerHeaderTimeout time.Duration `env:"SERVER_HEADER_TIMEOUT" envDefault:"60s"`

	// Database
	DBConnectionStr string `env:"DB_CONNECTION_STR" envDefault:"postgresql://postgres:postgres@localhost:5432/taskflow?sslmode=disable"`

	// JWT
	JWTSecret     string        `env:"JWT_SECRET" envDefault:"change-me-in-production"`
	JWTExpiration time.Duration `env:"JWT_EXPIRATION" envDefault:"24h"`
}

var appConfig *Config

func GetConfig() *Config {
	if appConfig != nil {
		return appConfig
	}

	if err := godotenv.Load(".env"); err != nil {
		fmt.Println("No .env file found, reading from environment")
	}

	appConfig = &Config{}
	if err := env.Parse(appConfig); err != nil {
		panic(err)
	}

	return appConfig
}
