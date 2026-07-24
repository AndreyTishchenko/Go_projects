package config

import "github.com/ilyakaznacheev/cleanenv"

type Config struct {
	Port          string `env:"PORT" env-default:"8080"`
	DatabaseURL   string `env:"DATABASE_URL" env-required:"true"`
	AdminLogin    string `env:"ADMIN_LOGIN" env-required:"true"`
	AdminPassword string `env:"ADMIN_PASSWORD" env-required:"true"`
}

func Load() (Config, error) {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
