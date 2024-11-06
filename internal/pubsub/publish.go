package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	jsonData, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("coundn't marshal json: %w", err)
	}

	return ch.PublishWithContext(
		context.Background(),
		exchange, key, false, false,
		amqp.Publishing{
			ContentType: "application-json",
			Body:        jsonData,
		},
	)
}

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	var b bytes.Buffer
	encoder := gob.NewEncoder(&b)
	if err := encoder.Encode(val); err != nil {
		return fmt.Errorf("couldn't encode data to gob: %w", err)
	}
	gobData := b.Bytes()

	return ch.PublishWithContext(
		context.Background(),
		exchange,
		key,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/gob",
			Body:        gobData,
		},
	)
}
