package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	zmq "github.com/pebbe/zmq4"
)

func main() {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("GO ZMQ SIMPLE SUBSCRIBER")
	fmt.Println(strings.Repeat("=", 60))

	// Step 1: Create SUB socket
	fmt.Println("\n[Step 1] Creating SUB socket...")
	subscriber, err := zmq.NewSocket(zmq.SUB)
	if err != nil {
		log.Fatal(err)
	}
	defer subscriber.Close()
	fmt.Println("✓ SUB socket created")

	// Step 2: Connect
	address := "tcp://localhost:5555"
	fmt.Printf("\n[Step 2] Connecting to publisher at %s...\n", address)
	err = subscriber.Connect(address)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ Connected to %s\n", address)

	// Step 3: Subscribe (Filter)
	// Critical: You MUST set a subscription, otherwise you get nothing.
	// Empty string "" means subscribe to everything.
	fmt.Println("\n[Step 3] Setting subscription filter...")
	subscriber.SetSubscribe("") 
	fmt.Println("✓ Subscribed to ALL topics")

	// Step 4: Receive Loop
	fmt.Println("\n[Step 4] Waiting for messages...")
	fmt.Println(strings.Repeat("-", 60))

	msgCount := 0
	for {
		// Receive string message (blocking)
		msg, err := subscriber.Recv(0)
		if err != nil {
			log.Fatal(err)
		}
		msgCount++

		timestamp := time.Now().Format("15:04:05")
		fmt.Printf("[%s] Received message #%d: %s\n", timestamp, msgCount, msg)
	}
}

