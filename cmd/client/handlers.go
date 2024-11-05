package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func _handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) {
	return func(playingState routing.PlayingState) {
		defer fmt.Print("\n> ")

		gs.HandlePause(playingState)
	}
}

func handlerMove(gs *gamelogic.GameState) func(gamelogic.ArmyMove) {
	return func(armyMove gamelogic.ArmyMove) {
		defer fmt.Print("\n> ")

		gs.HandleMove(armyMove)
	}
}
