package config

import (
	"os"
)

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	JWTTTLHours int
}

func Load() Config {
	return Config{
		Port:        getenv("APP_PORT", "8080"),
		DatabaseURL: getenv("DATABASE_URL", "postgres://tarkovhelper:tarkovhelper@localhost:5432/tarkovhelper?sslmode=disable"),
		JWTSecret:   getenv("JWT_SECRET", "dev_super_secret_change_me"),
		JWTTTLHours: getenvInt("JWT_TTL_HOURS", 720),
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func getenvInt(k string, def int) int {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	n := 0
	sign := 1
	i := 0
	if len(v) > 0 && v[0] == '-' {
		sign = -1
		i = 1
	}
	for ; i < len(v); i++ {
		c := v[i]
		if c < '0' || c > '9' {
			return def
		}
		n = n*10 + int(c-'0')
	}
	return n * sign
}
