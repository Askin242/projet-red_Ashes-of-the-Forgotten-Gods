package main

import (
	"main/pkg/display"
	"main/pkg/ui"
)

func main() {
	ui.InitScreen()
	ui.GameStartFunc = display.StartGame
	ui.ShowMainMenu()
}
