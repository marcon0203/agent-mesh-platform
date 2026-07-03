package handler

import "github.com/agentmesh/shared/pkg/config"

// Config 是 gateway-service 的全部可配置项。优先级：环境变量 >
// config.yaml（服务目录下，或 CONFIG_FILE 指定的路径）> 这里给的默认值。
type Config struct {
	PostgresDSN           string `yaml:"postgres_dsn" env:"GATEWAY_POSTGRES_DSN"`
	RedisAddr             string `yaml:"redis_addr" env:"GATEWAY_REDIS_ADDR"`
	OrchestrationGRPCAddr string `yaml:"orchestration_grpc_addr" env:"ORCHESTRATION_GRPC_ADDR"`
	HTTPAddr              string `yaml:"http_addr"`
}

func LoadConfig() (Config, error) {
	cfg := Config{
		PostgresDSN:           "postgres://agentmesh:agentmesh@127.0.0.1:5432/agentmesh?sslmode=disable",
		RedisAddr:             "127.0.0.1:6379",
		OrchestrationGRPCAddr: "127.0.0.1:9090",
		HTTPAddr:              ":8080",
	}
	err := config.Load(config.Path("config.yaml"), &cfg)
	return cfg, err
}
