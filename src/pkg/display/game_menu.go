package display

import (
	"errors"
	"fmt"
	"main/pkg/save"

	"github.com/awesome-gocui/gocui"
)

var ErrReturnToMainMenu = errors.New("return to main menu")

var gameMenuOpen = false
var gameMenuSelected = 0

func showGameMenu(g *gocui.Gui, v *gocui.View) error {
	if gameMenuOpen {
		return nil
	}

	gameMenuOpen = true
	gameMenuSelected = 0
	maxX, maxY := g.Size()

	menuWidth := 40
	menuHeight := 12
	menuX := (maxX - menuWidth) / 2
	menuY := (maxY - menuHeight) / 2

	if menuView, err := g.SetView("game_menu", menuX, menuY, menuX+menuWidth, menuY+menuHeight, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		menuView.Frame = true
		menuView.Title = " Game Menu "
		updateGameMenuView(menuView)
	}

	g.SetCurrentView("game_menu")
	return nil
}

func updateGameMenuView(v *gocui.View) {
	v.Clear()
	fmt.Fprintln(v, "")

	menuItems := []string{
		"Save & Continue",
		"Save & Return to Main Menu",
		"Save & Quit Game",
		"Cancel",
	}

	for i, item := range menuItems {
		if i == gameMenuSelected {
			fmt.Fprintf(v, "  \033[7m[%d] %s\033[0m\n", i+1, item)
		} else {
			fmt.Fprintf(v, "  [%d] %s\n", i+1, item)
		}
	}

	fmt.Fprintln(v, "")
	fmt.Fprintln(v, "  Use arrows/numbers, Enter/ESC")
}

func closeGameMenu(g *gocui.Gui) error {
	gameMenuOpen = false
	g.DeleteView("game_menu")
	g.SetCurrentView("game")
	return nil
}

func handleGameMenuChoice(g *gocui.Gui, choice int) error {
	switch choice {
	case 1: // Save & Continue
		save.SaveAny("player", gameState.player)
		return closeGameMenu(g)

	case 2: // Save & Return to Main Menu
		save.SaveAny("player", gameState.player)
		return ErrReturnToMainMenu

	case 3: // Save & Quit Game
		save.SaveAny("player", gameState.player)
		return gocui.ErrQuit

	case 4: // Cancel
		return closeGameMenu(g)

	default:
		return nil
	}
}

func setupGameMenuKeybindings(g *gocui.Gui) error {
	// Game menu (ESC key)
	if err := g.SetKeybinding("", gocui.KeyEsc, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if gameMenuOpen {
			return closeGameMenu(g)
		}
		return showGameMenu(g, v)
	}); err != nil {
		return err
	}

	// Arrow key navigation
	if err := g.SetKeybinding("game_menu", gocui.KeyArrowUp, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if gameMenuSelected > 0 {
			gameMenuSelected--
			updateGameMenuView(v)
		}
		return nil
	}); err != nil {
		return err
	}

	if err := g.SetKeybinding("game_menu", gocui.KeyArrowDown, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if gameMenuSelected < 3 {
			gameMenuSelected++
			updateGameMenuView(v)
		}
		return nil
	}); err != nil {
		return err
	}

	// Enter key to select
	if err := g.SetKeybinding("game_menu", gocui.KeyEnter, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return handleGameMenuChoice(g, gameMenuSelected+1)
	}); err != nil {
		return err
	}

	// Number key shortcuts
	if err := g.SetKeybinding("game_menu", '1', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return handleGameMenuChoice(g, 1)
	}); err != nil {
		return err
	}
	if err := g.SetKeybinding("game_menu", '2', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return handleGameMenuChoice(g, 2)
	}); err != nil {
		return err
	}
	if err := g.SetKeybinding("game_menu", '3', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return handleGameMenuChoice(g, 3)
	}); err != nil {
		return err
	}
	if err := g.SetKeybinding("game_menu", '4', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return handleGameMenuChoice(g, 4)
	}); err != nil {
		return err
	}
	if err := g.SetKeybinding("game_menu", gocui.KeyEsc, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return closeGameMenu(g)
	}); err != nil {
		return err
	}

	return nil
}
