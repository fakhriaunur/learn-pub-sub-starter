package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	jsonData, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("coundn't marshal json: %w", err)
	}

	err = ch.PublishWithContext(context.Background(),
		exchange, key, false, false,
		amqp.Publishing{
			ContentType: "application-json",
			Body:        jsonData,
		})
	if err != nil {
		return fmt.Errorf("couldn't publish with context: %w", err)
	}

	return nil
}
