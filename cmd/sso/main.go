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
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	ssov1.UnimplementedAuthServer
}

func main() {
	cfg := core_config.NewConfigMust()
	time.Local = cfg.TimeZone

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

	log.Debug("application time zone", zap.Any("zone", time.Local))

	lis, err := net.Listen("tcp", ":44044")
	if err != nil {
		log.Fatal(err)
	}

	g := grpc.NewServer()
	ssov1.RegisterAuthServer(g, &server{})
	reflection.Register(g)

	log.Println("sso port :44044")
	log.Fatal(g.Serve(lis))
}
