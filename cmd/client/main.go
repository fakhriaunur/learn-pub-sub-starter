package main

import (
	"fmt"
	"log"
	"strconv"

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
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+gameState.GetUsername(),
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
		handlerPause(gameState),
	); err != nil {
		log.Fatalf("couldn't subscribe to army moves: %v", err)
	}

	if err := pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+gameState.GetUsername(),
		routing.ArmyMovesPrefix+".*",
		pubsub.SimpleQueueTransient,
		handlerMove(publishCh, gameState),
	); err != nil {
		log.Fatalf("couldn't subscribe to army moves: %v", err)
	}

	if err := pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		routing.WarRecognitionsPrefix,
		routing.WarRecognitionsPrefix+".#",
		pubsub.SimpleQueueDurable,
		handlerWar(publishCh, gameState),
	); err != nil {
		log.Fatalf("couldn't subscribe to war recognitions: %v", err)
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
				routing.ArmyMovesPrefix+"."+mv.Player.Username,
				mv,
			); err != nil {
				fmt.Printf("error: %v", err)
				continue
			}
			fmt.Printf("Moved %v units to %s\n", len(mv.Units), mv.ToLocation)

		case "status":
			gameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			if len(words) != 2 {
				fmt.Printf("expecting two arguments: spam <number>\n")
				continue
			}

			numOfSpam, err := strconv.Atoi(words[1])
			if err != nil {
				fmt.Printf("coudln't parse int: %v\n", err)
				continue
			}

			for i := 1; i <= numOfSpam; i++ {
				maliciousLogMsg := gamelogic.GetMaliciousLog()

				if err := pubsub.PublishGob(
					publishCh,
					routing.ExchangePerilTopic,
					routing.GameLogSlug+".*",
					maliciousLogMsg,
				); err != nil {
					fmt.Printf("error: %v\n", err)
					continue
				}
				fmt.Printf("Publishing message #%d\n", i)
			}

			// fmt.Println("Spamming not allowed yet!")
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
