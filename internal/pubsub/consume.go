package pubsub

import (
	"bytes"
	"encoding/gob"
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
	return subscribe[T](
		conn,
		exchange,
		queueName,
		key,
		simpleQueueType,
		handler,
		func(data []byte) (T, error) {
			var target T
			err := json.Unmarshal(data, &target)
			return target, err
		},
	)
}

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) AckType,
) error {
	return subscribe[T](
		conn,
		exchange,
		queueName,
		key,
		simpleQueueType,
		handler,
		func(data []byte) (T, error) {
			buffer := bytes.NewBuffer(data)
			decoder := gob.NewDecoder(buffer)
			var target T
			err := decoder.Decode(&target)
			return target, err
		},
	)
}

func subscribe[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) AckType,
	decoder func([]byte) (T, error),
) error {
	ch, queue, err := DeclareAndBind(
		conn,
		exchange,
		queueName,
		key,
		simpleQueueType,
	)
	if err != nil {
		return fmt.Errorf("couldn't declare and bind: %w", err)
	}

	if err := ch.Qos(10, 0, false); err != nil {
		return fmt.Errorf("couldn't set qos: %w", err)
	}

	msgs, err := ch.Consume(
		queue.Name,
		"",
		false, false, false, false,
		amqp.Table{
			"x-dead-letter-exchange": "peril_dlx",
		},
	)
	if err != nil {
		return fmt.Errorf("couldn't consume: %w", err)
	}

	go func() {
		defer ch.Close()
		for msg := range msgs {
			target, err := decoder(msg.Body)
			if err != nil {
				fmt.Printf("couldn't decode: %n\n", err)
			}

			ackType := handler(target)

			switch ackType {
			case Ack:
				msg.Ack(true)
			case NackRequeue:
				msg.Nack(false, true)
			case NackDiscard:
				msg.Nack(false, false)
			}
		}
	}()

	return nil
}
