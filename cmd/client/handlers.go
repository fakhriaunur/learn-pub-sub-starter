package main

import (
	"fmt"
	"time"

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
				gamelogic.RecognitionOfWar{
					Attacker: armyMove.Player,
					Defender: gs.GetPlayerSnap(),
				},
			); err != nil {
				fmt.Printf("couldn't publish: %v", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		}
		fmt.Println("error: unknown move outcome")
		return pubsub.NackDiscard
	}
}

func handlerWar(ch *amqp.Channel, gs *gamelogic.GameState) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(recogWar gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")

		fmt.Printf("RecogWar: %+v\n", recogWar)
		warOutcome, winner, loser := gs.HandleWar(recogWar)
		fmt.Printf("Winner: %s, Loser: %s\n", winner, loser)
		fmt.Printf("Current username: %s\n", gs.GetUsername())

		var ackType pubsub.AckType
		var message string
		switch warOutcome {
		case gamelogic.WarOutcomeNotInvolved:
			ackType = pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			ackType = pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon,
			gamelogic.WarOutcomeYouWon:
			message = fmt.Sprintf("%s won a war against %s", winner, loser)
			ackType = pubsub.Ack
		case gamelogic.WarOutcomeDraw:
			message = fmt.Sprintf("A war between %s and %s resulted in a draw", winner, loser)
			ackType = pubsub.Ack
		default:
			fmt.Println("unknown war outcome")
		}

		gameLog := routing.GameLog{
			CurrentTime: time.Now(),
			Username:    gs.GetUsername(),
			Message:     message,
		}

		if ackType == pubsub.Ack {
			if err := pubsub.PublishGob(
				ch,
				routing.ExchangePerilTopic,
				routing.GameLogSlug+"."+gs.GetUsername(),
				gameLog,
			); err != nil {
				fmt.Printf("coudln't publish log: %v", err)
				return pubsub.NackRequeue
			}
		}
		return ackType
	}
}
