package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	zmq "github.com/pebbe/zmq4"
)

func main() {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("GO ZMQ TOPIC SUBSCRIBER")
	fmt.Println(strings.Repeat("=", 60))

	// Get topic filter from command line arguments
	filter := ""
	if len(os.Args) > 1 {
		filter = os.Args[1]
	}

	subscriber, _ := zmq.NewSocket(zmq.SUB)
	defer subscriber.Close()
	subscriber.Connect("tcp://localhost:5556")

	// Set subscription filter
	subscriber.SetSubscribe(filter)

	if filter == "" {
		fmt.Println("✓ Subscribed to ALL topics")
	} else {
		fmt.Printf("✓ Subscribed to filter: '%s'\n", filter)
	}

	fmt.Println("\nWaiting for messages...")
	fmt.Println(strings.Repeat("-", 60))

	msgCount := 0
	for {
		msg, err := subscriber.Recv(0)
		if err != nil {
			log.Fatal(err)
		}
		msgCount++

		timestamp := time.Now().Format("15:04:05")
		fmt.Printf("[%s] #%3d %s\n", timestamp, msgCount, msg)
	}
}

