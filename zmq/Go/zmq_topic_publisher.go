package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	zmq "github.com/pebbe/zmq4"
)

func main() {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("GO ZMQ TOPIC PUBLISHER")
	fmt.Println(strings.Repeat("=", 60))

	publisher, _ := zmq.NewSocket(zmq.PUB)
	defer publisher.Close()
	publisher.Bind("tcp://*:5556")

	fmt.Println("\n✓ Publisher started on port 5556")
	time.Sleep(2 * time.Second)
	fmt.Println("Starting to publish...\n")

	topics := []string{"Temperature", "Humidity", "Pressure"}
	msgCount := 0

	for {
		msgCount++
		// Randomly select a topic
		topic := topics[rand.Intn(len(topics))]
		var value int
		var unit string

		switch topic {
		case "Temperature":
			value = rand.Intn(16) + 15
			unit = "°C"
		case "Humidity":
			value = rand.Intn(41) + 40
			unit = "%"
		case "Pressure":
			value = rand.Intn(41) + 980
			unit = "hPa"
		}

		// Construct message: "Topic ValueUnit"
		// e.g., "Temperature 25°C"
		msg := fmt.Sprintf("%s %d%s", topic, value, unit)

		publisher.Send(msg, 0)

		timestamp := time.Now().Format("15:04:05")
		// Formatting output similar to Python's ljust
		fmt.Printf("[%s] #%3d [%-12s] %d%s\n", timestamp, msgCount, topic, value, unit)

		time.Sleep(500 * time.Millisecond)
	}
}

