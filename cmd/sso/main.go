package main

import (
	"log/slog"

	"github.com/pchek28/sso/internal/app"
	"github.com/pchek28/sso/internal/config"
	logger "github.com/pchek28/sso/internal/lib/logger"
)

func main() {
	cfg := config.MustLoadConfig()

	log := logger.SetupLogger(cfg.Env)

	log.Info("starting application", slog.Any("config", cfg))

	application := app.New(log, cfg.GRPC.Port, cfg.StoragePath, cfg.TokenTTL)

	application.MustRun()

}
