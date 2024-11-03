package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
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

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("couldn't open the channle: %v", err)
	}

	if err := pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey,
		routing.PlayingState{
			IsPaused: true,
		}); err != nil {
		log.Fatalf("couldn't publish json: %v", err)
	}

	fmt.Println("Successfully connected to the URL")

	// wait for ctrl + c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	interruptSignal := <-signalChan

	fmt.Printf("\n%s signal received\n", interruptSignal.String())
	fmt.Println("Peril server is shutting down...")
}
