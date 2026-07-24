package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Port              string        `env:"PORT" env-default:"8080"`
	DatabaseURL       string        `env:"DATABASE_URL" env-required:"true"`
	AdminLogin        string        `env:"ADMIN_LOGIN" env-required:"true"`
	AdminPasswordHash string        `env:"ADMIN_PASSWORD_HASH" env-required:"true"`
	SessionTTL        time.Duration `env:"SESSION_TTL" env-default:"24h"`
}

func Load() (Config, error) {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return Config{}, err
	}
	if cfg.SessionTTL <= 0 {
		return Config{}, fmt.Errorf("SESSION_TTL must be greater than zero")
	}
	return cfg, nil
}
