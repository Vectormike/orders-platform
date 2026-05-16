package config

import "os"

type Config struct {
	Env         string
	Port        string
	GinMode     string
	DatabaseURL string
}

func Load() Config {
	env := getEnv("APP_ENV", "dev")

	return Config{
		Env:         env,
		Port:        getEnv("PORT", "8080"),
		GinMode:     resolveGinMode(env),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}
}

func resolveGinMode(env string) string {
	switch env {
	case "prod":
		return "release"
	case "test":
		return "test"
	default:
		return "debug"
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
