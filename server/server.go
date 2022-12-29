package main

import (
	proto "Passivereplication/grpc"
	"flag"
	"log"
	"net"
	"strconv"

	"google.golang.org/grpc"
)

type Server struct {
	proto.UnimplementedIncrementServiceServer
	name  string
	port  int
	value int32
}

var port = flag.Int("port", 0, "server port number") // create the port that recieves the port that the client wants to access to

func main() {

	flag.Parse()

	server := &Server{
		name: "serverName",
		port: *port,
	}

	go startServer(server)

	for {

	}
}

func startServer(server *Server) {
	grpcServer := grpc.NewServer()                                           // create a new grpc server
	listen, err := net.Listen("tcp", "localhost:"+strconv.Itoa(server.port)) // creates the listener

	if err != nil {
		log.Fatalln("Could not start listener")
	}

	log.Printf("Server started at port %v", server.port)

	proto.RegisterIncrementServiceServer(grpcServer, server) // register the server
	serverError := grpcServer.Serve(listen)

	if serverError != nil {
		log.Printf("Could not register server")
	}
}

// func (c *Server) Increment(ctx context.Context, in *proto.IncRequest) (*proto.IncResponse, error){
// 	return nil,nil

// 	if in.Amount <= 0 {
// 		log.Println("return fail")
// 		return &proto.IncResponse{}, errors.New("You cmust increment!")
// 	} else {
// 		//c.newAmount = c.newAmount + in.Amount
// 	}

// 	return &proto.Ack{Ack: fail}, errors.New("Something is very wrong")
// }
