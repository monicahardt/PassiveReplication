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
	port   int
	leader *proto.IncrementServiceClient
	servers []Server
	amount int32
}

type Server struct {
	server proto.IncrementServiceClient
	isLeader           bool
	port int32
}

//var port = flag.Int("port", 0, "server port number") // create the port that recieves the port that the client wants to access to

func newFrontend() *Frontend{
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

func (f *Frontend) connectToServer(portNumber int32){
	//dialing the server
	conn, err := grpc.Dial("localhost:"+strconv.Itoa(int(portNumber)), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Could not connect: %s", err)
	}
	//if nothing is wrong
	log.Printf("Frontend connected to server at port: %v\n", portNumber)

	newServerToAdd := proto.NewIncrementServiceClient(conn)
	isLeader, err := newServerToAdd.GetLeaderRequest(context.Background(),&proto.Empty{})

	f.servers = append(f.servers, Server{
		server: newServerToAdd,
		isLeader: isLeader.IsLeader,
		port: portNumber,
	})

	if(isLeader.IsLeader == true){
		f.leader = &newServerToAdd
		fmt.Println("found the leader")
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
	fmt.Printf("Printing the number of serevrs the frontend is connected to, %v", len(f.servers))
	leader := *f.leader
	response, er:= leader.Increment(ctx, in)
	if(er != nil){
		fmt.Println("The frontend found out that the leader is dead, now going to find the new one")
		//here we want to remove the dead leader form the frontends slice of servers
		
		for i := 0; i < len(f.servers); i++ {
			if(5001 == f.servers[i].port){
				f.servers = removeServer(f.servers,i)
			}
		}
	}

	for i := 0; i < len(f.servers); i++ {
		fmt.Println("Going though slice to find new leader")
		message, _ := f.servers[i].server.GetLeaderRequest(context.Background(),&proto.Empty{})
		if(message.IsLeader){
			f.leader = &f.servers[i].server
			fmt.Println("Updated the leeeeeader WHOOOOOOOOO")
		} else {
			fmt.Println("YOOOOOOOOOOO")
		}
	}
	leaderNew := *f.leader
	response, err := leaderNew.Increment(ctx, in)
	return response, err
}

func removeServer(s []Server, i int) []Server {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}