package core_config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Location struct{ *time.Location }

func (l *Location) Decode(value string) error {
	zone, err := time.LoadLocation(value)
	if err != nil {
		return fmt.Errorf("load time zone: %q: %w", value, err)
	}

	l.Location = zone
	return nil
}

type Config struct {
	TimeZone        *time.Location `envconfig:"TIME_ZONE" default:"UTC"`
	GRPCPort        string         `envconfig:"GRPC_PORT" default:"44044"`
	Issuer          string         `envconfig:"JWT_ISSUER" default:"sso"`
	PrivateKeyPath  string         `envconfig:"JWT_PRIVATE_KEY_PATH" default:"keys/private.pem"`
	AccessTTL       time.Duration  `envconfig:"ACCESS_TTL" default:"15m"`
	RefreshTTL      time.Duration  `envconfig:"REFRESH_TTL" default:"720h"`
	ShutdownTimeout time.Duration  `envconfig:"SHUTDOWN_TIMEOUT" default:"10s"`
}

func NewConfig() (*Config, error) {
	var config Config

	if err := envconfig.Process("", &config); err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	return &config, nil
}

func NewConfigMust() *Config {
	cfg, err := NewConfig()
	if err != nil {
		err = fmt.Errorf("get env keys %w", err)
		panic(err)
	}
	return cfg
}
