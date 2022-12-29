package main

import (
	proto "Passivereplication/grpc"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"google.golang.org/grpc"
)

type Server struct {
	proto.UnimplementedIncrementServiceServer
	port          int
	value         int32
	isLeader      bool
	serverclients []ReplicaClient
}

type ReplicaClient struct {
	replicaClient proto.IncrementServiceClient
	isLeader      bool
	port          int
	value         int32
}

var (
	port          = flag.Int("port", 0, "server port number") // create the port that recieves the port that the client wants to access to
	localAddress  int32
	leaderAddress int32
)

func main() {
	flag.Parse()
	//Here we create a new server
	//if the server has port 5001 it is chosen to be the first leader
	s := &Server{
		port:          *port,
		serverclients: make([]ReplicaClient, 0),
		isLeader:      *port == 5001,
	}

	if s.isLeader {
		fmt.Println("now started the leader")
	}

	//starting this server and listening on the port
	go startServer(s)

	fmt.Println("Connecting to other replication nodes")
	go s.connectToReplica(s, 5001)
	go s.connectToReplica(s, 5002)
	go s.connectToReplica(s, 5003)
	fmt.Println("Finished connection to the other replicas")

	go func() {
		for {
			time.Sleep(30000 * time.Millisecond)
			s.heartbeat()
		}
	}()

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

	proto.RegisterIncrementServiceServer(grpcServer, server)
	serverError := grpcServer.Serve(listen)

	if serverError != nil {
		log.Printf("Could not register server")
	}
}

func (server *Server) GetLeaderRequest(_ context.Context, _ *proto.Empty) (*proto.LeaderMessage, error) {
	return &proto.LeaderMessage{Id: int32(server.port), IsLeader: server.isLeader}, nil
}

func (s *Server) connectToReplica(server *Server, portNumber int32) {
	// Connect with replica's IP
	conn, err := grpc.Dial("localhost:"+strconv.Itoa(int(portNumber)), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect gRPC server :: %v", err)
	}
	defer conn.Close()

	// Create a new replicationClient struct to be stored in the replication server struct
	newReplicationClient := proto.NewIncrementServiceClient(conn)

	// Check if the replica manager is the leader
	// Since the server may not exist yet, we keep asking until we get a response
	var newIsLeader bool
	//var newId int32
	for {
		replicationClientMessage, err := newReplicationClient.GetLeaderRequest(context.Background(), &proto.Empty{})
		if err == nil {
			newIsLeader = replicationClientMessage.IsLeader

			break
		}
		// Retry until the connection is established
		time.Sleep(500 * time.Millisecond)
	}

	log.Printf("Successfully connected to the replica with port %v!", portNumber)

	// Append the new ReplicationClient to the list of replicationClients stored in the ReplicationServer struct
	server.serverclients = append(server.serverclients, ReplicaClient{
		replicaClient: newReplicationClient,
		isLeader:      newIsLeader,
		port:          int(portNumber),
	})

	//fmt.Printf("**************printing size: %v", len(server.serverclients))

	// Keep the go-routine running to not close the connection
	for {

	}
}

func (c *Server) Increment(ctx context.Context, in *proto.IncRequest) (*proto.IncResponse, error) {
	fmt.Printf("the value is: %v", c.value)
	if in.Amount <= 0 {
		log.Println("return fail")
		return &proto.IncResponse{}, errors.New("You must increment!")
	}
	if c.isLeader {
		//update it self first
		c.value = c.value + in.Amount
		fmt.Printf("the value after increment is: %v", c.value)
		fmt.Println("Increment in the leader class was called")

		for i := 0; i < len(c.serverclients); i++ {
			// fmt.Println("in loooooooooppppppp")
			// fmt.Println("trying to update the other replicas")
			_, err := c.serverclients[i].replicaClient.Replicate(context.Background(), &proto.ReplicationValue{Value: c.value})

			if err != nil {
				fmt.Println("Failed to update a replica")
			}
		}
	} else {
		fmt.Printf("increment in non-leader was called the value is not: %v", c.value)
	}

	fmt.Println("udated all servers")
	return &proto.IncResponse{NewAmount: c.value}, nil
}

func (s *Server) Replicate(ctx context.Context, in *proto.ReplicationValue) (*proto.ReplicationAck, error) {
	//If the replicated server does not have the correct value
	fmt.Printf("Replicate was called the value before was %v", s.value)
	if s.value != in.Value {
		s.value = in.Value
	}
	fmt.Printf("Replicate was called the value before was %v", s.value)
	return &proto.ReplicationAck{}, nil
}

// BULLY!!!!!!!
func (s *Server) heartbeat() {
	fmt.Println("A heartbeat was sent from the leader")
	// For each replication client, make sure it does still exist
	for i := 0; i < len(s.serverclients); i++ {
		// Request info from the client repeatedly to discover changes and to notice of the connection is lost
		isLeader, err := s.serverclients[i].replicaClient.GetLeaderRequest(context.Background(), &proto.Empty{})

		if err != nil {
			// If an error has occurred it means we were unable to reach the replication node.
			log.Println("Unable to reach a replication node!")

			// Check if the unreachable node was the leader
			lostNodeWasLeader := s.serverclients[i].isLeader

			// Remove the node
			log.Println("Now removing lost node from known replication nodes")
			s.serverclients = removeReplicationClient(s.serverclients, i)
			log.Printf("System now has %v replication nodes\n", len(s.serverclients))

			if lostNodeWasLeader {
				log.Println("Unable to reach leader node")
				log.Println("A new leader must be assigned")

				// We now need to assign a new leader
				// Each replication node has a unique ID. Leadership is passed to the node with the highest ID
				// Since we already know the ID of every replication node, we simply assign ourselves as the leader
				// if we have the highest id with no need to contact the other nodes first.
				// Next time the other nodes ping us they will learn that we are the new leader.

				// Determine the node with the highest ID
				var highest int32
				for j := 0; j < len(s.serverclients); j++ {
					if int32(s.serverclients[j].port) > highest {
						highest = int32(s.serverclients[j].port)
					}
				}
				// If we have the highest index set as new leader
				if highest == int32(s.port) {
					log.Println("I am the new leader!")
					s.isLeader = true
				}
				log.Println("A new leader has been picked")
				log.Printf("the server with port: %v, has boolean isLeader = %v ", s.port, s.isLeader)
			}
		} else {
			// Set isLeader based on node info. This lets the node discover new leaders
			s.serverclients[i].isLeader = isLeader.IsLeader
		}
	}
}

func removeReplicationClient(s []ReplicaClient, i int) []ReplicaClient {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

// go run server/server.go -port 5001
