package main

import (
	proto "Passivereplication/grpc"
	"context"
	"fmt"
	"log"
	"net"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Frontend struct {
	proto.UnimplementedIncrementServiceServer
	//name   string
<<<<<<< HEAD
	port    int
	leader  proto.IncrementServiceClient
=======
	port   int
	leader *proto.IncrementServiceClient
>>>>>>> f9418df0ebd692d5599fe499194f13d1b31cb085
	servers []Server
	amount  int32
}

type Server struct {
	server   proto.IncrementServiceClient
	isLeader bool
}

//var port = flag.Int("port", 0, "server port number") // create the port that recieves the port that the client wants to access to

func newFrontend() *Frontend {
	fmt.Printf("called frontend method")

	frontend := &Frontend{
		//name:            "frontend",
		//port:            *port,
		servers: make([]Server, 0),
	}

	//go startFrontend(frontend)
	//we have to make servers at all these ports at program start
	go frontend.connectToServer(5001)
	go frontend.connectToServer(5002)
	go frontend.connectToServer(5003)
	fmt.Println("try returning the new frontend with connnections")
	return frontend
}

func (f *Frontend) connectToServer(portNumber int32) {
	//dialing the server
	conn, err := grpc.Dial("localhost:"+strconv.Itoa(int(portNumber)), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Could not connect: %s", err)
	}
	//if nothing is wrong
	log.Printf("Frontend connected to server at port: %v\n", portNumber)

	newServerToAdd := proto.NewIncrementServiceClient(conn)
	isLeader, err := newServerToAdd.GetLeaderRequest(context.Background(), &proto.Empty{})

	f.servers = append(f.servers, Server{
		server:   newServerToAdd,
		isLeader: isLeader.IsLeader,
	})

<<<<<<< HEAD
	if isLeader.IsLeader == true {
		f.leader = newServerToAdd
=======
	if(isLeader.IsLeader == true){
		f.leader = &newServerToAdd
		fmt.Println("found the leader")
>>>>>>> f9418df0ebd692d5599fe499194f13d1b31cb085
	}
	//defer conn.Close()
	wait := make(chan bool)
	<-wait

	// for {
	// fmt.Println("Running frontend")
	// }
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

func (f *Frontend) Increment(ctx context.Context, in *proto.IncRequest) (*proto.IncResponse, error) {
<<<<<<< HEAD
	response, err := f.leader.Increment(ctx, in)
	return response, err
}
=======
	fmt.Println("somethiong went wrong in frontend increment method")
	// if(f.leader == nil){
	// 	fmt.Println("We are trying to increment on a leader that is null")
	// }

	
	// leader := *f.leader
	// response, err:= leader.Increment(ctx, in)
	// return response, err
	return &proto.IncResponse{NewAmount: 5}, nil
}
>>>>>>> f9418df0ebd692d5599fe499194f13d1b31cb085
