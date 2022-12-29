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
	clientPort   = flag.Int("cPort", 0, "client port number")
)

var frontend *Frontend

func main() {
	flag.Parse()	
	frontend = newFrontend()


	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		amountToIncrement, _ := strconv.ParseInt(scanner.Text(), 10, 0)
		fmt.Printf("Scanned an input, %v", amountToIncrement)
		_, err := frontend.Increment(context.Background(), &proto.IncRequest{Amount: int32(amountToIncrement)})
		if err != nil {
			fmt.Printf("Increment went wrong")
		}
	}
	for {
		
	}
}

// go run client/client.go -cPort 4040



