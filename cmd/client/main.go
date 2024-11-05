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
	fmt.Println("Starting Peril client...")

	const rabbitMQConnStr = "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(rabbitMQConnStr)
	if err != nil {
		log.Fatalf("couldn't dial the url: %v", err)
	}
	defer conn.Close()
	fmt.Println("Peril game client connected to RabbitMQ")

	publishCh, err := conn.Channel()
	if err != nil {
		log.Fatalf("couldn't open channel: %v", err)
	}

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("couldn't retrieve the username: %v", err)
	}

	_, queue, err := pubsub.DeclareAndBind(
		conn, routing.ExchangePerilDirect,
		routing.PauseKey+"."+username,
		routing.PauseKey, pubsub.SimpleQueueTransient,
	)
	if err != nil {
		log.Fatalf("couldn't declare and bind: %v", err)
	}
	fmt.Printf("Queue %v declared and bound\n", queue.Name)

	gameState := gamelogic.NewGameState(username)

	if err := pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+username,
		routing.ArmyMovesPrefix+".*",
		pubsub.SimpleQueueTransient,
		handlerMove(gameState),
	); err != nil {
		log.Fatalf("couldn't subscribe to army moves: %v", err)
	}

	for {
		words := gamelogic.GetInput()

		if len(words) == 0 {
			continue
		}
		switch words[0] {
		case "spawn":
			if err := gameState.CommandSpawn(words); err != nil {
				fmt.Printf("couldn't spawn: %v", err)
				continue
			}
		case "move":
			mv, err := gameState.CommandMove(words)
			if err != nil {
				fmt.Printf("couldn't move: %v", err)
				continue
			}

			if err := pubsub.PublishJSON(
				publishCh,
				routing.ExchangePerilTopic,
				routing.ArmyMovesPrefix+"."+username,
				mv,
			); err != nil {
				fmt.Printf("error: %v", err)
				continue
			}
			log.Println("The move was published successfully!")

		case "status":
			gameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			log.Println("Quitting the game...")
			log.Println("Peril client is shutting down...")
			return

		default:
			fmt.Println("unknown command")
		}

	}
}
