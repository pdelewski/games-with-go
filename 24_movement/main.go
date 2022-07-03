package main

// https://youtu.be/Jy919y3ezOI?t=1346

import (
	"github.com/pdelewski/autotel/rtlib"
	"github.com/pdelewski/games-with-go/24_movement/game"
	"github.com/pdelewski/games-with-go/24_movement/ui2d"
)

func main() {
	rtlib.SumoAutoInstrument()
	ui := &ui2d.UI2d{}
	game.Run(ui)
}
