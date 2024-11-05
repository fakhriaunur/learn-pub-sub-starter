package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(playingState routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")

		gs.HandlePause(playingState)

		return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(armyMove gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")

		moveOutcome := gs.HandleMove(armyMove)
		fmt.Printf("moveOutcome: %v\n", moveOutcome)
		if moveOutcome == gamelogic.MoveOutcomeSafe ||
			moveOutcome == gamelogic.MoveOutcomeMakeWar {
			return pubsub.Ack
		}

		return pubsub.NackDiscard
	}
}
