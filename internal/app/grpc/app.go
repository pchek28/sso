package app_grpc

import (
	"fmt"
	"log/slog"
	"net"

	authgrpc "github.com/pchek28/sso/internal/grpc/auth"
	"google.golang.org/grpc"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}

func New(
	log *slog.Logger,
	port int,
) *App {
	gRPCServer := grpc.NewServer()

	authgrpc.Register(gRPCServer)

	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		a.log.Error("failed to run gRPC server", slog.String("error", err.Error()))
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "app_grpc.App.Run"

	log := a.log.With(
		slog.String("op", op),
		slog.Int("port", a.port),
	)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		log.Error("failed to listen", slog.String("error", err.Error()))
		return fmt.Errorf("%s: failed to listen: %w", op, err)
	}

	log.Info("starting gRPC server", slog.String("addr", l.Addr().String()))

	if err := a.gRPCServer.Serve(l); err != nil {
		log.Error("failed to serve gRPC server", slog.String("error", err.Error()))
		return fmt.Errorf("%s: failed to serve gRPC server: %w", op, err)
	}

	return nil
}

func (a *App) Stop() {
	const op = "app_grpc.App.Stop"

	log := a.log.With(
		slog.String("op", op),
	)
	log.Info("stopping gRPC server")
	a.gRPCServer.GracefulStop()
	log.Info("gRPC server stopped")
}
