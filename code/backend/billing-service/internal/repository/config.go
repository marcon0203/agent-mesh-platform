package repository

import "github.com/agentmesh/shared/pkg/config"

// Config 是 billing-service 的全部可配置项。优先级：环境变量 >
// config.yaml（服务目录下，或 CONFIG_FILE 指定的路径）> 这里给的默认值。
type Config struct {
	PostgresDSN string `yaml:"postgres_dsn" env:"BILLING_POSTGRES_DSN"`
	RedisAddr   string `yaml:"redis_addr" env:"BILLING_REDIS_ADDR"`
	HTTPAddr    string `yaml:"http_addr"`
}

func LoadConfig() (Config, error) {
	cfg := Config{
		PostgresDSN: "postgres://agentmesh:agentmesh@127.0.0.1:5432/agentmesh?sslmode=disable",
		RedisAddr:   "127.0.0.1:6379",
		HTTPAddr:    ":8083",
	}
	err := config.Load(config.Path("config.yaml"), &cfg)
	return cfg, err
}
