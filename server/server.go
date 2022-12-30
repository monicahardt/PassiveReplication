package main

import (
	proto "Passivereplication/grpc"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
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
	//setting the server log file
	f, err := os.OpenFile("log.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	
	flag.Parse()
	//Here we create a new server
	//if the server has port 5001 it is chosen to be the first leader
	s := &Server{
		port:          *port,
		serverclients: make([]ReplicaClient, 0),
		isLeader: *port == 5003,
	}

	if(s.isLeader){
		fmt.Println("The leader has started")
	}

	//starting this server and listening on the port
	go startServer(s)
	go s.connectToReplica(s,5001)
	go s.connectToReplica(s,5002)
	go s.connectToReplica(s,5003)


	go func() {
		for {
			time.Sleep(30000 * time.Millisecond)
			s.heartbeat()
		}
	}()

	for {
	}
}


//The normal startserver method. Creates new grpc server and creates the listener
func startServer(server *Server) {
	grpcServer := grpc.NewServer()                                           // create a new grpc server
	listen, err := net.Listen("tcp", "localhost:"+strconv.Itoa(server.port)) // creates the listener

	if err != nil {
		log.Fatalln("Could not start listener")
	}

	log.Printf("Server started at port %v\n", server.port)

	proto.RegisterIncrementServiceServer(grpcServer, server)
	serverError := grpcServer.Serve(listen)

	if serverError != nil {
		log.Printf("Could not register server")
	}
}

	// Connect with replica's IP
func (s *Server) connectToReplica(server *Server, portNumber int32) {
	conn, err := grpc.Dial("localhost:"+strconv.Itoa(int(portNumber)), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect gRPC server :: %v", err)
	}
	defer conn.Close()

	// Create a new replicationClient struct to be stored in the replication server struct
	newReplicationClient := proto.NewIncrementServiceClient(conn)

	// Check if the newReplicationClient is the leader
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
	fmt.Printf("Successfully connected to the replica with port %v!\n",portNumber)
	log.Printf("Server with id: %v connected to replicaclient with id: %v\n",s.port, portNumber)

	// Append the new ReplicationClient to the list of replicationClients stored in the ReplicationServer struct
	server.serverclients = append(server.serverclients, ReplicaClient{
		replicaClient: newReplicationClient,
		isLeader:      newIsLeader,
		port:          int(portNumber),
	})

	//If this is removed we have som goroutines issues and all servers will have 0 servers in their slice by the time we want to incremt!
	fmt.Printf("This server has %v clients\n", len(server.serverclients))

	// Keep the go-routine running to not close the connection
	for {
	}
}


//tells us if this server is the leader
func (server *Server) GetLeaderRequest(_ context.Context, _ *proto.Empty) (*proto.LeaderMessage, error) {
	//log.Printf("It is %v that server with id: %v is the leader\n", server.isLeader, server.port)
	//fmt.Printf("It is %v that server with id: %v is the leader\n", server.isLeader, server.port)
	return &proto.LeaderMessage{Id: int32(server.port), IsLeader: server.isLeader}, nil
}

func (c *Server) Increment(ctx context.Context, in *proto.IncRequest) (*proto.IncResponse, error){
	//checking if the amount is valid
	if in.Amount <= 0 {
		log.Println("return fail")
		return &proto.IncResponse{}, errors.New("You must increment!")
	} 

	if(c.isLeader){
		//update it self first
		log.Printf("A client wants to increment value with: %v\n", in.Amount)
		c.value = c.value + in.Amount
		log.Printf("The leader has now incremented to: %v\n", c.value)
		
		//after updating itself, we have to notify the other clients
		//with raft this is done with a heartbeat method
		for i := 0; i < len(c.serverclients); i++ {
			//calling replicate with the value of the leader
			_, err := c.serverclients[i].replicaClient.Replicate(context.Background(),&proto.ReplicationValue{Value: c.value})
			if(err != nil){
				log.Println("Failed to update a replica")
			}
		}
	}
	return &proto.IncResponse{NewAmount: c.value, Id: int32(c.port)}, nil
} 

func (s *Server) Replicate(ctx context.Context, in *proto.ReplicationValue) (*proto.ReplicationAck, error) {
	//If the replicated server does not have the correct value
	if(s.value != in.Value){
		s.value = in.Value
	}
	log.Printf("Server with id: %v now has updated value: %v\n", s.port, s.value)
	return &proto.ReplicationAck{}, nil
}


// BULLY!!!!!!!
func (s *Server) heartbeat() {
	// For each replication client, make sure it does still exist
	for i := 0; i < len(s.serverclients); i++ {
		// Request info from the client repeatedly to discover changes and to notice of the connection is lost
		isLeader, err := s.serverclients[i].replicaClient.GetLeaderRequest(context.Background(), &proto.Empty{})

		if err != nil {
			// If an error has occurred it means we were unable to reach the replication node.
			fmt.Printf("Server with id: %v was unable to reach a replicationclient!\n", s.port)

			// Check if the unreachable node was the leader
			lostNodeWasLeader := s.serverclients[i].isLeader

			// Remove the node
			//log.Println("Now removing lost node from known replication nodes")
			s.serverclients = removeReplicationClient(s.serverclients, i)

			if lostNodeWasLeader {
				log.Printf("Server with id: %v found that the leader has crashed\n", s.port)
				log.Println("A new leader must be assigned")

				// We now need to assign a new leader
				//each replicationclient has a port. We assign the leader role
				//to the client with the highest portnumber
				// Next time the other nodes ping us they will learn that we are the new leader.

				// Determine the node with the highest ID/port
				var highest int32
				for j := 0; j < len(s.serverclients); j++ {
					if int32(s.serverclients[j].port) > highest {
						highest = int32(s.serverclients[j].port)
					}
				}
				// If we have the highest index set as new leader
				if highest == int32(s.port) {
					s.isLeader = true
					log.Printf("Server with id: %v is now the new new leader!\n", s.port)
				}
				//log.Println("A new leader has been picked")
				//log.Printf("the server with port: %v, has boolean isLeader = %v ", s.port, s.isLeader)
			} else {
				log.Printf("Server with id: %v found that a replicaclient has crashed\n", s.port)
			}
		} else {
			// Set isLeader based on node info. This lets the node discover new leaders
			s.serverclients[i].isLeader = isLeader.IsLeader
			//fmt.Println("There was no error in increment at the leader")
		}
	}
}

func removeReplicationClient(s []ReplicaClient, i int) []ReplicaClient {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

// go run server/server.go -port 5001
