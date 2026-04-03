package config

import (
	"errors"
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	App   AppConfig
	DB    DBConfig
	JWT   JWTConfig
	Minio MinioConfig
	AI    AIConfig
}
type AIConfig struct {
	AnthropicKey string `env:"ANTHROPIC_API_KEY" env-default:""`
	OpenAIKey    string `env:"OPENAI_API_KEY"    env-default:""`
}

type AppConfig struct {
	Host           string `env:"APP_HOST"          env-default:"0.0.0.0"`
	Port           string `env:"APP_PORT"          env-default:"8080"`
	Env            string `env:"APP_ENV"           env-default:"local"`
	SwaggerHost    string `env:"SWAGGER_HOST"      env-default:"localhost:8080"`
	AllowedOrigins string `env:"ALLOWED_ORIGINS"   env-default:"*"`
}

type DBConfig struct {
	Host     string `env:"DB_HOST"     env-default:"localhost"`
	Port     string `env:"DB_PORT"     env-default:"5432"`
	User     string `env:"DB_USER"     env-required:"true"`
	Password string `env:"DB_PASSWORD" env-required:"true"`
	Name     string `env:"DB_NAME"     env-required:"true"`
	SSLMode  string `env:"DB_SSLMODE"  env-default:"disable"`
}

type JWTConfig struct {
	SecretKey       string `env:"JWT_SECRET"          env-required:"true"`
	AccessTokenTTL  int    `env:"JWT_ACCESS_TTL_MIN"  env-default:"15"`
	RefreshTokenTTL int    `env:"JWT_REFRESH_TTL_DAY" env-default:"30"`
}

type MinioConfig struct {
	Endpoint  string `env:"MINIO_ENDPOINT"   env-required:"true"`
	AccessKey string `env:"MINIO_ACCESS_KEY" env-required:"true"`
	SecretKey string `env:"MINIO_SECRET_KEY" env-required:"true"`
	Bucket    string `env:"MINIO_BUCKET"     env-default:"oqyrman"`
	UseSSL    bool   `env:"MINIO_USE_SSL"    env-default:"false"`
	PublicURL string `env:"MINIO_PUBLIC_URL" env-default:"https://api.oqyrman.app/minio"`
}

func New() (*Config, error) {
	cfg := &Config{}

	// Пробуем загрузить .env файл — если нет, читаем из переменных окружения
	// Это позволяет работать и локально (с .env) и в Docker (через env_file / -e)
	if err := cleanenv.ReadConfig(".env", cfg); err != nil {
		log.Println("no .env file found, reading from environment variables")
		if err := cleanenv.ReadEnv(cfg); err != nil {
			return nil, err
		}
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate проверяет критические настройки при старте приложения.
func (cfg *Config) Validate() error {
	if len([]byte(cfg.JWT.SecretKey)) < 32 {
		return errors.New("JWT_SECRET must be at least 32 bytes")
	}
	if cfg.App.Env == "production" && cfg.App.AllowedOrigins == "*" {
		return errors.New("ALLOWED_ORIGINS must not be '*' in production")
	}
	if cfg.App.Env == "production" && cfg.DB.SSLMode == "disable" {
		return errors.New("DB_SSLMODE must not be 'disable' in production")
	}
	return nil
}
