package pubsub

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int,
) (*amqp.Channel, amqp.Queue, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("coudn't open the channel: %w", err)
	}

	queue, err := ch.QueueDeclare(
		queueName, simpleQueueType == 0, simpleQueueType == 1,
		simpleQueueType == 1, false, nil,
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
