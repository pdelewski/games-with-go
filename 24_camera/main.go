package main

// https://youtu.be/Jy919y3ezOI?t=1346

import (
	"github.com/pdelewski/games-with-go/24_camera/game"
	"github.com/pdelewski/games-with-go/24_camera/ui2d"
	"github.com/pdelewski/autotel/rtlib"
)

func main() {
	rtlib.SumoAutoInstrument()
	ui := &ui2d.UI2d{}
	game.Run(ui)
}
