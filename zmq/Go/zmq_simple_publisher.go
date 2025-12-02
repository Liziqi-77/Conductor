package main

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	zmq "github.com/pebbe/zmq4"
)

func main() {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("GO ZMQ SIMPLE PUBLISHER")
	fmt.Println(strings.Repeat("=", 60))

	// Step 1: Create PUB socket
	// Go handles context implicitly in NewSocket usually, or you can create one.
	// For simplicity, we use the default context.
	fmt.Println("\n[Step 1] Creating PUB socket...")
	publisher, err := zmq.NewSocket(zmq.PUB)
	if err != nil {
		log.Fatal(err)
	}
	// defer ensures the socket closes when main function exits
	defer publisher.Close()
	fmt.Println("✓ PUB socket created")

	// Step 2: Bind
	address := "tcp://*:5555"
	fmt.Printf("\n[Step 2] Binding to %s...\n", address)
	err = publisher.Bind(address)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ Publisher bound to %s\n", address)

	// Step 3: Wait
	fmt.Println("\n[Step 3] Waiting 2 seconds for subscribers to connect...")
	time.Sleep(2 * time.Second)
	fmt.Println("✓ Ready to publish!")

	// Step 4: Loop
	fmt.Println("\n[Step 4] Publishing messages (Press Ctrl+C to stop)...")
	fmt.Println(strings.Repeat("-", 60))

	msgCount := 0
	for {
		msgCount++
		// Generate random temperature
		temperature := rand.Intn(16) + 15 // 15-30
		msg := fmt.Sprintf("Temperature: %d°C", temperature)

		// Send message (0 means no special flags)
		_, err := publisher.Send(msg, 0)
		if err != nil {
			log.Println("Error sending:", err)
			break
		}

		// Log
		timestamp := time.Now().Format("15:04:05")
		fmt.Printf("[%s] Sent message #%d: %s\n", timestamp, msgCount, msg)

		time.Sleep(1 * time.Second)
	}
}

