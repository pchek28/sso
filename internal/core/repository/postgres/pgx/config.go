package core_pgx_pool

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Host     string        `envconfig:"HOST" required:"TRUE"`
	Port     string        `envconfig:"PORT" default:"5433"`
	User     string        `envconfig:"USER" required:"TRUE"`
	Password string        `envconfig:"PASSWORD" required:"TRUE"`
	Database string        `envconfig:"DB" required:"TRUE"`
	Timeout  time.Duration `envconfig:"TIMEOUT" required:"TRUE"`
}

func NewConfig() (Config, error) {
	var config Config

	if err := envconfig.Process("POSTGRES", &config); err != nil {
		return Config{}, fmt.Errorf("process envconfig: %w", err)
	}

	return config, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		err = fmt.Errorf("get Postgres connection pool config: %w", err)
		panic(err)
	}

	return config
}
