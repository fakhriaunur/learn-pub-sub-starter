package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")

	url := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Fatalf("couldn't dial the url: %v", err)
	}
	defer conn.Close()

	fmt.Println("Successfully connected to the URL")

	// wait for ctrl + c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	interruptSignal := <-signalChan

	fmt.Printf("\n%s signal received\n", interruptSignal.String())
	fmt.Println("Peril server is shutting down...")
}
