package core_config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	TimeZone        *time.Location
	GRPCPort        string
	Issuer          string
	PrivateKeyPath  string
	AccessTTL       time.Duration
	RefreshTTL      time.Duration
	ShutdownTimeout time.Duration
}

func NewConfig() (*Config, error) {
	tz := GetFromENV("TIME_ZONE", "UTC")

	zone, err := time.LoadLocation(tz)
	if err != nil {
		return nil, fmt.Errorf("load time zone: %s: %w", tz, err)
	}

	return &Config{
		TimeZone:        zone,
		GRPCPort:        GetFromENV("GRPC_PORT", "44044"),
		Issuer:          GetFromENV("JWT_ISSUER", "sso"),
		PrivateKeyPath:  GetFromENV("JWT_PRIVATE_KEY_PATH", "keys/private.pem"),
		AccessTTL:       15 * time.Minute,
		RefreshTTL:      720 * time.Hour,
		ShutdownTimeout: 10 * time.Second,
	}, nil
}

func NewConfigMust() *Config {
	cfg, err := NewConfig()
	if err != nil {
		err = fmt.Errorf("get env keys %w", err)
		panic(err)
	}
	return cfg
}

func GetFromENV(key string, def string) string {
	if data := os.Getenv(key); data != "" {
		return data
	}
	return def
}
