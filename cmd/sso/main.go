package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	ssov1 "github.com/pchek28/protos/gen/go/sso"
	core_config "github.com/pchek28/sso/internal/core/config"
	core_logger "github.com/pchek28/sso/internal/core/logger"
	core_pgx_pool "github.com/pchek28/sso/internal/core/repository/postgres/pgx"
	core_security "github.com/pchek28/sso/internal/core/security"
	auth_repository "github.com/pchek28/sso/internal/features/repository/postgres"
	auth_service "github.com/pchek28/sso/internal/features/service"
	auth_transport_grpc "github.com/pchek28/sso/internal/features/transport/grpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	ssov1.UnimplementedAuthServer
}

func main() {
	cfg := core_config.NewConfigMust()
	time.Local = cfg.TimeZone.Location

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM,
	)

	defer cancel()

	log, err := core_logger.NewLogger(core_logger.NewConfigMust())
	if err != nil {
		fmt.Println("failed to init application logger", err)
		os.Exit(1)
	}
	defer func() {
		_ = log.Sync()
		_ = log.Close()
	}()
	ctx = core_logger.ToContext(ctx, log)

	log.Debug("application time zone", zap.Any("zone", time.Local))

	pool, err := core_pgx_pool.NewPool(ctx, core_pgx_pool.NewConfigMust())
	if err != nil {
		log.Fatal("init postgres pool", zap.Error(err))
	}
	defer pool.Close()

	key, err := core_security.LoadEd25519Key(cfg.PrivateKeyPath)
	if err != nil {
		log.Fatal("load private key", zap.String("path", cfg.PrivateKeyPath), zap.Error(err))
	}
	issuer := core_security.NewJWTIssuer(key, cfg.Issuer, cfg.AccessTTL)

	authRepo := auth_repository.New(pool)
	authService, err := auth_service.New(authRepo, issuer, cfg.RefreshTTL)
	if err != nil {
		log.Fatal("init auth service", zap.Error(err))
	}

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatal("listen", zap.String("port", cfg.GRPCPort), zap.Error(err))
	}

	g := grpc.NewServer()
	ssov1.RegisterAuthServer(g, auth_transport_grpc.New(authService, log))
	reflection.Register(g)

	serveErr := make(chan error, 1)
	go func() { serveErr <- g.Serve(lis) }()
	log.Info("sso started", zap.String("port", cfg.GRPCPort))

	select {
	case <-ctx.Done():
		log.Info("shutdown signal received")
	case err := <-serveErr:
		log.Fatal("grpc serve", zap.Error(err))
	}

	stopped := make(chan struct{})
	go func() {
		g.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		log.Info("sso stopped gracefully")
	case <-time.After(cfg.ShutdownTimeout):
		log.Warn("graceful stop timed out, forcing")
		g.Stop()
	}
}
