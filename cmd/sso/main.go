package main

import (
	"log"
	"net"

	ssov1 "github.com/pchek28/protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	ssov1.UnimplementedAuthServer
}

func main() {
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
