package main

import (
	proto "Passivereplication/grpc"
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
)

type Client struct {
	id         int
	portNumber int
}

var (
	clientPort = flag.Int("cPort", 0, "client port number")
)

var frontend *Frontend

func main() {
<<<<<<< HEAD
	flag.Parse()
	frontend := newFrontend()
=======
	flag.Parse()	
	frontend = newFrontend()
>>>>>>> f9418df0ebd692d5599fe499194f13d1b31cb085

	client := &Client{
		id:         *clientPort,
		portNumber: *clientPort,
	}

	go scanInput(client, frontend)
	for {
		
	}
}

func scanInput(client *Client, frontend *Frontend) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		// if the client's input is bid, scans the amount that the client tries to bid and calls the bid() in the frontend
		// if the clietn's input is result, prints the highest bidder and highest bid
		// else prints invalid
		scanner.Scan()
		amountToIncrement, _ := strconv.ParseInt(scanner.Text(), 10, 0)
		fmt.Printf("Scanned an input, %v", amountToIncrement)
		reponse, err := frontend.Increment(context.Background(), &proto.IncRequest{Amount: int32(amountToIncrement)})

		if err != nil {
			fmt.Printf("Increment went wrong")
		}

		fmt.Printf("The new value is %v", reponse.NewAmount)
	}
}

// go run client/frontend.go client/client.go -cPort 8082
