package pubsub

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AckType int

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
)

type SimpleQueueType int

const (
	SimpleQueueDurable SimpleQueueType = iota
	SimpleQueueTransient
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
) (*amqp.Channel, amqp.Queue, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("coudn't open the channel: %w", err)
	}

	table := amqp.Table{
		"x-dead-letter-exchange": "peril_dlx",
	}

	queue, err := ch.QueueDeclare(
		queueName, simpleQueueType == SimpleQueueDurable, simpleQueueType == SimpleQueueTransient,
		simpleQueueType == SimpleQueueTransient, false,
		table,
	)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("couldn't declare the queue: %w", err)
	}

	if err := ch.QueueBind(
		queueName, key, exchange,
		false, nil,
	); err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("couldn't bind the queue: %w", err)
	}

	return ch, queue, nil
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) AckType,
) error {
	ch, queue, err := DeclareAndBind(
		conn, exchange, queueName,
		key, simpleQueueType,
	)
	if err != nil {
		return fmt.Errorf("couldnt declare and bind: %w", err)
	}

	deliveryCh, err := ch.Consume(
		queue.Name, "", false, false, false, false,
		amqp.Table{},
	)
	if err != nil {
		return fmt.Errorf("couldn't consume: %w", err)
	}

	go func() {
		defer ch.Close()
		for msg := range deliveryCh {
			var jsonMsg T
			if err := json.Unmarshal(msg.Body, &jsonMsg); err != nil {
				fmt.Printf("couldn't unmarshal: %v", err)
			}

			ackType := handler(jsonMsg)
			switch ackType {
			case Ack:
				if err := msg.Ack(false); err != nil {
					fmt.Printf("couldn't deliver acknowledge: %v", err)
				}
				fmt.Println("Ack")
			case NackRequeue:
				if err := msg.Nack(false, true); err != nil {
					fmt.Printf("coudln't deliver acknowledge: %v", err)
				}
				fmt.Println("NackRequeue")
			case NackDiscard:
				if err := msg.Nack(false, false); err != nil {
					fmt.Printf("couldn't deliver acknowledge: %v", err)
				}
				fmt.Println("NackDiscard")
			}
		}
	}()

	return nil
}
