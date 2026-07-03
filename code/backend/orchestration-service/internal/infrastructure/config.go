package infrastructure

import "github.com/agentmesh/shared/pkg/config"

// Config 是 orchestration-service 的全部可配置项。优先级：环境变量 >
// config.yaml（服务目录下，或 CONFIG_FILE 指定的路径）> 这里给的默认值。
type Config struct {
	PostgresDSN         string `yaml:"postgres_dsn" env:"ORCHESTRATION_POSTGRES_DSN"`
	RedisAddr           string `yaml:"redis_addr" env:"ORCHESTRATION_REDIS_ADDR"`
	GRPCAddr            string `yaml:"grpc_addr"`
	AdminHTTPAddr       string `yaml:"admin_http_addr"`
	ModelProviderEncKey string `yaml:"model_provider_enc_key" env:"MODEL_PROVIDER_ENC_KEY"`
}

func LoadConfig() (Config, error) {
	cfg := Config{
		PostgresDSN:         "postgres://agentmesh:agentmesh@127.0.0.1:5432/agentmesh?sslmode=disable",
		RedisAddr:           "127.0.0.1:6379",
		GRPCAddr:            ":9090",
		AdminHTTPAddr:       ":8082",
		ModelProviderEncKey: "000102030405060708090a0b0c0d0e0f",
	}
	err := config.Load(config.Path("config.yaml"), &cfg)
	return cfg, err
}
