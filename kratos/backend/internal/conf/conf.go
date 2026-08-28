package conf

// 配置加载：config.yaml + 环境变量覆盖（DB_DSN/DB_DSN_REPLICA/REDIS_ADDR/
// OPENSEARCH_ADDR/JWT_SECRET/PORT/GRPC_PORT）。

import (
	"os"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
)

func Load(path string) (*Bootstrap, error) {
	c := config.New(config.WithSource(file.NewSource(path)))
	if err := c.Load(); err != nil {
		return nil, err
	}
	var bc Bootstrap
	if err := c.Scan(&bc); err != nil {
		return nil, err
	}
	if v := os.Getenv("DB_DSN"); v != "" {
		bc.Data.DbDsn = v
	}
	if v := os.Getenv("DB_DSN_REPLICA"); v != "" {
		bc.Data.DbDsnReplica = v
	}
	if v := os.Getenv("REDIS_ADDR"); v != "" {
		bc.Data.RedisAddr = v
	}
	if v := os.Getenv("OPENSEARCH_ADDR"); v != "" {
		bc.Data.OpensearchAddr = v
	}
	if v := os.Getenv("JWT_SECRET"); v != "" {
		bc.Auth.JwtSecret = v
	}
	if v := os.Getenv("PAYMENT_ENCRYPT_KEY"); v != "" {
		bc.Payment.EncryptKey = v
	}
	if v := os.Getenv("STRIPE_SECRET_KEY"); v != "" {
		bc.Payment.StripeSecretKey = v
	}
	if v := os.Getenv("STRIPE_WEBHOOK_SECRET"); v != "" {
		bc.Payment.StripeWebhookSecret = v
	}
	if v := os.Getenv("NP_API_KEY"); v != "" {
		bc.Payment.NpApiKey = v
	}
	if v := os.Getenv("NP_IPN_SECRET"); v != "" {
		bc.Payment.NpIpnSecret = v
	}
	if v := os.Getenv("PORT"); v != "" {
		bc.Server.HttpAddr = ":" + v
	}
	if v := os.Getenv("GRPC_PORT"); v != "" {
		bc.Server.GrpcAddr = ":" + v
	}
	return &bc, nil
}
