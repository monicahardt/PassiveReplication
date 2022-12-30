package main

import (
	proto "Passivereplication/grpc"
	"context"
	"fmt"
	"log"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Frontend struct {
	proto.UnimplementedIncrementServiceServer
	port   int
	leader *proto.IncrementServiceClient
	servers []Server
	amount  int32
}

type Server struct {

	server   proto.IncrementServiceClient
	isLeader bool
	port     int32

}


func newFrontend() *Frontend{
	frontend := &Frontend{
		servers: make([]Server, 0),
	}

	//we have to make servers at all these ports at program start
	go frontend.connectToServer(5001)
	go frontend.connectToServer(5002)
	go frontend.connectToServer(5003)
	return frontend
}

func (f *Frontend) connectToServer(portNumber int32) {

	//dialing the server
	conn, err := grpc.Dial("localhost:"+strconv.Itoa(int(portNumber)), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Could not connect: %s\n", err)
	}
	//if nothing is wrong
	log.Printf("Frontend connected to server at port: %v\n", portNumber)

	newServerToAdd := proto.NewIncrementServiceClient(conn)
	isLeader, err := newServerToAdd.GetLeaderRequest(context.Background(), &proto.Empty{})

	f.servers = append(f.servers, Server{
		server:   newServerToAdd,
		isLeader: isLeader.IsLeader,
		port: isLeader.Id,
	})

	if isLeader.IsLeader == true {
		f.leader = &newServerToAdd
		fmt.Printf("Fronted set the leader to be server with port: %v", isLeader.Id)
	}

	//defer conn.Close()
	wait := make(chan bool)
	<-wait
}


func (f *Frontend) Increment(ctx context.Context, in *proto.IncRequest) (*proto.IncResponse, error) {
	leader := *f.leader
	response, err:= leader.Increment(ctx, in)
	
	if(err != nil){
		fmt.Println("The frontend found out that the leader is dead")
		//find highest portnumber an remove it. This is hardcoding, but don't know what else to do
		toRemove := 0
		highestPort := int32(5001)
		for i := 0; i < len(f.servers); i++ {
			if(f.servers[i].port > highestPort){
				highestPort = f.servers[i].port
				toRemove = i
			}
		}
		f.servers = removeServer(f.servers, toRemove)
		//finding the new leader
		for i := 0; i < len(f.servers); i++ {
			message, _ := f.servers[i].server.GetLeaderRequest(context.Background(),&proto.Empty{})
			if(message.IsLeader){
				f.leader = &f.servers[i].server
				fmt.Printf("Frontend updated the leader to be server at port: %v\n", message.Id)
			} 
		}
		leaderNew := *f.leader
		response, err = leaderNew.Increment(ctx, in)
	}
	return response, err
}

func removeServer(s []Server, i int) []Server {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]

}
