package config

import (
	"fmt"
	"log"
	"os"
)

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type RedisConfig struct {
	Host string
	Port string
}

func (db DBConfig) ConnString() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		db.Host, db.Port, db.User, db.Password, db.Name,
	)
}

type Config struct {
	DB      DBConfig
	Redis   RedisConfig
	AppPort string
}

func Load() Config {
	return Config{
		DB: DBConfig{
			Host:     mustGetEnv("POSTGRES_HOST"),
			Port:     mustGetEnv("POSTGRES_PORT"),
			User:     mustGetEnv("POSTGRES_USER"),
			Password: mustGetEnv("POSTGRES_PASSWORD"),
			Name:     mustGetEnv("POSTGRES_DB"),
		},
		Redis: RedisConfig{
			Host: mustGetEnv("REDIS_HOST"),
			Port: mustGetEnv("REDIS_PORT"),
		},
		AppPort: getEnv("APP_PORT", "8080"),
	}
}

func mustGetEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("environment variable %s is required", key)
	}
	return val
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
