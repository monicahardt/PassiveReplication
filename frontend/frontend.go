package main

import (
	proto "Passivereplication/grpc"
	"context"
	"log"
	"net"
	"strconv"

	"google.golang.org/grpc"
)

type Frontend struct {
	proto.UnimplementedIncrementServiceServer
	name   string
	port   int
	server proto.IncrementServiceClient
}

func (f *Frontend) Increment(ctx context.Context, in *proto.IncRequest) (*proto.IncResponse, error) {
	return nil, nil
}

func startFrontend(frontend *Frontend) {
	grpcServer := grpc.NewServer()
	listen, err := net.Listen("tcp", "localhost:"+strconv.Itoa(frontend.port))

	if err != nil {
		log.Fatalln("Could not start listener")
	}

	log.Printf("Frontend started at port %v", frontend.port)

	proto.RegisterIncrementServiceServer(grpcServer, frontend)
	serverError := grpcServer.Serve(listen)

	if serverError != nil {
		log.Printf("Could not register frontend")
	}
}
