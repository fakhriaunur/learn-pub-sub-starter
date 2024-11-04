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
	fmt.Println("Starting Peril client...")

	const rabbitMQConnStr = "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(rabbitMQConnStr)
	if err != nil {
		log.Fatalf("couldn't dial the url: %v", err)
	}
	defer conn.Close()
	fmt.Println("Peril game client connected to RabbitMQ")

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("couldn't retrieve the username: %v", err)
	}

	_, queue, err := pubsub.DeclareAndBind(
		conn, routing.ExchangePerilDirect,
		fmt.Sprintf("%s.%s", routing.PauseKey, username),
		routing.PauseKey, 1,
	)
	if err != nil {
		log.Fatalf("couldn't declare and bind: %v", err)
	}
	fmt.Printf("Queue %v declared and bound\n", queue.Name)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	interruptSignal := <-signalChan

	fmt.Printf("\n%s signal received\n", interruptSignal.String())
	fmt.Println("Peril client is shutting down...")

	// ch, err := conn.Channel()
	// if err != nil {
	// 	log.Fatalf("couldn't open the channel: %v", err)
	// }

}
