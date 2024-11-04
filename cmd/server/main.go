package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")

	const rabbitMQConnStr = "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(rabbitMQConnStr)
	if err != nil {
		log.Fatalf("couldn't dial the url: %v", err)
	}
	defer conn.Close()
	fmt.Println("Peril game server connected to RabbitMQ")

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("couldn't open the channel: %v", err)
	}

	gamelogic.PrintServerHelp()
	for {
		words := gamelogic.GetInput()

		if len(words) == 0 {
			continue
		}
		if words[0] == "pause" {
			if err := pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey,
				routing.PlayingState{
					IsPaused: true,
				}); err != nil {
				log.Fatalf("couldn't publish json: %v", err)
			}
			fmt.Println("Pause message sent!")

		} else if words[0] == "resume" {
			if err := pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey,
				routing.PlayingState{
					IsPaused: false,
				}); err != nil {
				log.Fatalf("couldn't publish json: %v", err)
			}
			fmt.Println("Resume message sent!")

		} else if words[0] == "quit" {
			fmt.Println("Exiting the game...")
			break

		} else {
			fmt.Println("couldn't understand the command")
		}
	}

	// wait for ctrl + c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	interruptSignal := <-signalChan

	fmt.Printf("\n%s signal received\n", interruptSignal.String())
	fmt.Println("Peril server is shutting down...")
}
