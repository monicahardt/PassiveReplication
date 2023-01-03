package main

import (
	proto "Passivereplication/grpc"
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	portNumber int
	proto.UnimplementedDistributedDictionaryServer
	leader  *proto.DistributedDictionaryClient
	servers []Server
	//amount  int32
}

type Server struct {
	server   proto.DistributedDictionaryClient
	isLeader bool
	port     int32
}

var (
	clientPort = flag.Int("cPort", 0, "client port number")
)

func main() {

	flag.Parse()

	client := &Client{
		servers:    make([]Server, 0),
		portNumber: *clientPort,
	}

	go client.connectToServer(5001)
	go client.connectToServer(5002)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {

		//this is the input of the client who want to add or read
		input := scanner.Text()
		if input == "add" {
			scanner.Scan()
			word := scanner.Text()

			scanner.Scan()
			def := scanner.Text()

			response, err := client.Add(context.Background(), &proto.AddReq{Word: word, Def: def})

			if err != nil {
				fmt.Printf("Something is wrong with add \n")
			}
			fmt.Printf("The addition is %v", response.GetAccepted())

		} else if input == "read" {

			scanner.Scan()
			word := scanner.Text()

			def, err := client.Read(context.Background(), &proto.ReadReq{Word: word})

			if err != nil {
				fmt.Printf("Something is wrong with add \n")
			}
			if def.GetDef() == "" {
				fmt.Println("You requested definition for a word not in the dictionary")
			}
			fmt.Printf("The read word: %v, got definition: %v", word, def.GetDef())

		} else {
			fmt.Println("Invalid - cannot understand")
		}

	}
	for {

	}
}

func (c *Client) connectToServer(portNumber int32) {

	//dialing the server
	conn, err := grpc.Dial("localhost:"+strconv.Itoa(int(portNumber)), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Could not connect: %s\n", err)
	}
	//if nothing is wrong
	log.Printf("Frontend connected to server at port: %v\n", portNumber)

	newServerToAdd := proto.NewDistributedDictionaryClient(conn)
	//proto.NewIncrementServiceClient(conn)
	isLeader, _ := newServerToAdd.GetLeaderRequest(context.Background(), &proto.Empty{})

	c.servers = append(c.servers, Server{
		server:   newServerToAdd,
		isLeader: isLeader.IsLeader,
		port:     isLeader.Id,
	})

	if isLeader.IsLeader {
		c.leader = &newServerToAdd
		log.Printf("Fronted set the leader to be server with port: %v\n", isLeader.Id)
	}

	//defer conn.Close()
	wait := make(chan bool)
	<-wait
}

func (c *Client) Add(ctx context.Context, in *proto.AddReq) (*proto.AddRes, error) {
	leader := *c.leader
	response, err := leader.Add(ctx, in)
	if err != nil {
		log.Println("The client found out that the leader is dead")
		//find highest portnumber an remove it. This is hardcoding, but don't know what else to do
		toRemove := 0
		highestPort := int32(5001)
		for i := 0; i < len(c.servers); i++ {
			if c.servers[i].port > highestPort {
				highestPort = c.servers[i].port
				toRemove = i
			}
		}
		c.servers = removeServer(c.servers, toRemove)
		//finding the new leader
		for i := 0; i < len(c.servers); i++ {
			message, _ := c.servers[i].server.GetLeaderRequest(context.Background(), &proto.Empty{})
			if message.IsLeader {
				c.leader = &c.servers[i].server
				fmt.Printf("Frontend updated the leader to be server at port: %v\n", message.Id)
			}
		}
		leaderNew := *c.leader
		response, err = leaderNew.Add(ctx, in)
	}
	return response, err
}

func (c *Client) Read(ctx context.Context, in *proto.ReadReq) (*proto.ReadRes, error) {
	leader := *c.leader
	response, err := leader.Read(ctx, in)
	if err != nil {
		log.Println("The client found out that the leader is dead")
		//find highest portnumber an remove it. This is hardcoding, but don't know what else to do
		toRemove := 0
		highestPort := int32(5001)
		for i := 0; i < len(c.servers); i++ {
			if c.servers[i].port > highestPort {
				highestPort = c.servers[i].port
				toRemove = i
			}
		}
		c.servers = removeServer(c.servers, toRemove)
		//finding the new leader
		for i := 0; i < len(c.servers); i++ {
			message, _ := c.servers[i].server.GetLeaderRequest(context.Background(), &proto.Empty{})
			if message.IsLeader {
				c.leader = &c.servers[i].server
				fmt.Printf("Frontend updated the leader to be server at port: %v\n", message.Id)
			}
		}
		leaderNew := *c.leader
		response, err = leaderNew.Read(ctx, in)
	}
	return response, err
}

// func (c *Client) Increment(ctx context.Context, in *proto.IncRequest) (*proto.IncResponse, error) {
// 	leader := *c.leader
// 	response, err := leader.Increment(ctx, in)

// 	if err != nil {
// 		log.Println("The frontend found out that the leader is dead")
// 		//find highest portnumber an remove it. This is hardcoding, but don't know what else to do
// 		toRemove := 0
// 		highestPort := int32(5001)
// 		for i := 0; i < len(c.servers); i++ {
// 			if c.servers[i].port > highestPort {
// 				highestPort = c.servers[i].port
// 				toRemove = i
// 			}
// 		}
// 		c.servers = removeServer(c.servers, toRemove)
// 		//finding the new leader
// 		for i := 0; i < len(c.servers); i++ {
// 			message, _ := c.servers[i].server.GetLeaderRequest(context.Background(), &proto.Empty{})
// 			if message.IsLeader {
// 				c.leader = &c.servers[i].server
// 				fmt.Printf("Frontend updated the leader to be server at port: %v\n", message.Id)
// 			}
// 		}
// 		leaderNew := *c.leader
// 		response, err = leaderNew.Increment(ctx, in)
// 	}
// 	return response, err
// }

func removeServer(s []Server, i int) []Server {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]

}

// go run client/frontend.go client/client.go -cPort 8082
