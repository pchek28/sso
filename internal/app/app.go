package app

import (
	"log/slog"
	"time"

	app_grpc "github.com/pchek28/sso/internal/app/grpc"
)

type App struct {
	gRPCServer *app_grpc.App
}

func New(
	log *slog.Logger,
	gRPCPort int,
	storagePath string,
	tokenTTL time.Duration,
) *App {

	grpcApp := app_grpc.New(log, gRPCPort)

	return &App{
		gRPCServer: grpcApp,
	}
}

func (a *App) MustRun() {
	a.gRPCServer.MustRun()
}
