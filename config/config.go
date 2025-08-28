package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	// Config -.
	Config struct {
		App    `yaml:"app"`
		HTTP   `yaml:"http"`
		Log    `yaml:"logger"`
		PG     `yaml:"postgres"`
		ApiKey `yaml:"api_key"`
		JWT    `yaml:"jwt"`
		// OpenAI `yaml:"openai"`
	}

	// App -.
	App struct {
		Name    string `env-required:"true" yaml:"name"    env:"APP_NAME"`
		Version string `env-required:"true" yaml:"version" env:"APP_VERSION"`
	}

	// HTTP -.
	HTTP struct {
		Port string `env-required:"true" yaml:"port" env:"HTTP_PORT"`
	}

	// Log -.
	Log struct {
		Level string `env-required:"true" yaml:"log_level"   env:"LOG_LEVEL"`
	}

	// PG -.
	PG struct {
		PoolMax int    `env-required:"true" yaml:"pool_max" env:"PG_POOL_MAX"`
		URL     string `env-required:"true"                 env:"PG_URL"`
	}

	// ApiKey -.
	ApiKey struct {
		Key string `env-required:"true" yaml:"key" env:"API_KEY"`
	}

	// JWT -.
	JWT struct {
		Secret string `env-required:"true" yaml:"secret" env:"JWT_SECRET"`
	}

	// // OpenAI -.
	// OpenAI struct {
	// 	ApiKey      string `env-required:"true" yaml:"api_key" env:"OPENAI_API_KEY"`
	// 	AssistantID string `env-required:"true" yaml:"assistant_id" env:"OPENAI_ASSISTANT_ID"`
	// }
)

// NewConfig returns app config.
func NewConfig() (*Config, error) {
	cfg := &Config{}

	err := cleanenv.ReadConfig("./config/config.yml", cfg)
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	err = cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
