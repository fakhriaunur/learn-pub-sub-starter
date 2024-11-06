package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func handlerLog() func(routing.GameLog) pubsub.AckType {
	return func(gameLog routing.GameLog) pubsub.AckType {
		defer fmt.Print("> ")

		if err := gamelogic.WriteLog(gameLog); err != nil {
			fmt.Printf("couldn't write log: %v\n", err)
			return pubsub.NackRequeue
		}

		return pubsub.Ack
	}
}
