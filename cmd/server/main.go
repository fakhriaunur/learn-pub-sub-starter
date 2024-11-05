package main

import (
	"fmt"
	"log"

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

	_, queue, err := pubsub.DeclareAndBind(
		conn, routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug+".*",
		pubsub.SimpleQueueDurable,
	)
	if err != nil {
		log.Fatalf("couldn't declare and bind: %v", err)
	}
	fmt.Printf("%s successfully declared and bound\n", queue.Name)

	gamelogic.PrintServerHelp()
	for {
		words := gamelogic.GetInput()

		if len(words) == 0 {
			continue
		}
		switch words[0] {
		case "pause":
			fmt.Println("Publishing paused game state")
			if err := pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey,
				routing.PlayingState{
					IsPaused: true,
				}); err != nil {
				log.Fatalf("couldn't publish time: %v", err)
			}
			fmt.Println("Pause message sent!")

		case "resume":
			fmt.Println("Publishing resumed game state")
			if err := pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey,
				routing.PlayingState{
					IsPaused: false,
				}); err != nil {
				log.Fatalf("couldn't publish time: %v", err)
			}
			fmt.Println("Resume message sent!")

		case "quit":
			log.Println("Exiting the game...")
			log.Println("Peril client is shutting down...")
			return

		default:
			fmt.Println("couldn't understand the command")
		}
	}
}
