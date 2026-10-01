package phase0

import (
	"fmt"
	"net/url"
	"strings"
)

type Config struct {
	DatabaseURL, RedisAddr, S3Endpoint, S3PublicEndpoint, S3AccessKey, S3SecretKey, S3Bucket string
	QdrantURL, OTelEndpoint, AdminUsername, AdminPassword                                    string
	RequireRAG                                                                               bool
}

func LoadConfig(getenv func(string) string) (Config, error) {
	c := Config{
		DatabaseURL: getenv("ORBIS_DATABASE_URL"), RedisAddr: getenv("ORBIS_REDIS_ADDR"),
		S3Endpoint: getenv("ORBIS_S3_ENDPOINT"), S3PublicEndpoint: getenv("ORBIS_S3_PUBLIC_ENDPOINT"),
		S3AccessKey: getenv("ORBIS_S3_ACCESS_KEY"), S3SecretKey: getenv("ORBIS_S3_SECRET_KEY"), S3Bucket: getenv("ORBIS_S3_BUCKET"),
		QdrantURL: getenv("ORBIS_QDRANT_URL"), OTelEndpoint: getenv("ORBIS_OTEL_ENDPOINT"),
		AdminUsername: getenv("ORBIS_ADMIN_USERNAME"), AdminPassword: getenv("ORBIS_ADMIN_PASSWORD"),
	}
	for k, v := range map[string]string{"ORBIS_DATABASE_URL": c.DatabaseURL, "ORBIS_REDIS_ADDR": c.RedisAddr, "ORBIS_S3_ENDPOINT": c.S3Endpoint, "ORBIS_S3_ACCESS_KEY": c.S3AccessKey, "ORBIS_S3_SECRET_KEY": c.S3SecretKey, "ORBIS_S3_BUCKET": c.S3Bucket} {
		if strings.TrimSpace(v) == "" {
			return Config{}, fmt.Errorf("%s is required", k)
		}
	}
	if u, err := url.Parse(c.DatabaseURL); err != nil || u.Scheme == "" || u.Host == "" {
		return Config{}, fmt.Errorf("ORBIS_DATABASE_URL is invalid")
	}
	switch strings.ToLower(getenv("ORBIS_REQUIRE_RAG")) {
	case "", "false", "0":
	case "true", "1":
		c.RequireRAG = true
	default:
		return Config{}, fmt.Errorf("ORBIS_REQUIRE_RAG must be true or false")
	}
	if c.RequireRAG && strings.TrimSpace(c.QdrantURL) == "" {
		return Config{}, fmt.Errorf("ORBIS_QDRANT_URL is required when RAG is enabled")
	}
	return c, nil
}
