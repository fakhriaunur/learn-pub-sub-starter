package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(playingState routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")

		gs.HandlePause(playingState)

		return pubsub.Ack
	}
}

func handlerMove(ch *amqp.Channel, gs *gamelogic.GameState) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(armyMove gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")

		moveOutcome := gs.HandleMove(armyMove)
		fmt.Printf("moveOutcome: %v\n", moveOutcome)
		if moveOutcome == gamelogic.MoveOutcomeSafe {
			return pubsub.Ack
		}

		if moveOutcome == gamelogic.MoveOutcomeMakeWar {
			if err := pubsub.PublishJSON(
				ch,
				routing.ExchangePerilTopic,
				routing.WarRecognitionsPrefix+"."+gs.GetUsername(),
				armyMove,
			); err != nil {
				fmt.Printf("couldn't publish: %v", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		}

		return pubsub.NackDiscard
	}
}

func handlerWar(gs *gamelogic.GameState) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(recogWar gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")

		warOutcome, _, _ := gs.HandleWar(recogWar)

		switch warOutcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon,
			gamelogic.WarOutcomeYouWon,
			gamelogic.WarOutcomeDraw:
			return pubsub.Ack
		}

		fmt.Println("unknown war outcome")
		return pubsub.NackDiscard
	}
}
