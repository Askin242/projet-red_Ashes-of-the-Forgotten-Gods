package main

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/awesome-gocui/gocui"
)

var (
	username string
	race     string
	seed     string
	g        *gocui.Gui
	hovered  string
	errMsg   string
	dialog   bool
)

type MenuItem struct {
	text   string
	action func(*gocui.Gui, *gocui.View) error
}

var menuItems = []MenuItem{
	{"[1] New game file", showNewGameForm},
	{"[2] Load game file", loadGameFile},
	{"[3] Erase game file", eraseGameFile},
	{"[4] Exit", quit},
}

var selected = 0
var inForm = false
var fields = []string{"username", "race", "seed"}
var fieldIndex = 0
var races = []string{"Human", "Elf", "Dwarf"}
var raceIndex = 0

func setTerminalSize(cols, rows int) {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/C", fmt.Sprintf("mode con: cols=%d lines=%d", cols, rows))
		cmd.Run()
	} else {
		fmt.Printf("\033[8;%d;%dt", rows, cols)
	}
}

func centerText(text string, width int) string {
	lines := strings.Split(text, "\n")
	var centeredLines []string

	for _, line := range lines {
		spaces := (width - len(line)) / 2
		centeredLines = append(centeredLines, strings.Repeat(" ", spaces)+line)
	}

	return strings.Join(centeredLines, "\n")
}

func isLatinOnly(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < 'A' || (c > 'Z' && c < 'a') || c > 'z' {
			return false
		}
	}
	return true
}

func checkInput(g *gocui.Gui) {
	errMsg = ""
	v, _ := g.View("username")
	user := v.Buffer()
	v, _ = g.View("seed")
	seed := v.Buffer()

	if user != "" && (!isLatinOnly(user) || (user[0] < 'A' || user[0] > 'Z')) {
		if !isLatinOnly(user) {
			errMsg = "Username must be letters only"
		} else {
			errMsg = "Username must start with capital letter"
		}
		return
	}

	if seed != "" && !isLatinOnly(seed) {
		errMsg = "Seed must be letters only"
	}
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	updateHoverEffects(g)

	v, _ := g.SetView("logo", 0, 0, maxX-1, 18, 0)
	v.Frame = false
	v.Wrap = false

	logoPart1 := `                                                                           
     .oo        8                      .oPYo.  d'b   ooooo 8               
    .P 8        8                      8    8  8       8   8               
   .P  8 .oPYo. 8oPYo. .oPYo. .oPYo.   8    8 o8P      8   8oPYo. .oPYo.  
  oPooo8 Yb..   8    8 8oooo8 Yb..     8    8  8       8   8    8 8oooo8  
 .P    8   'Yb. 8    8 8.       'Yb.   8    8  8       8   8    8 8.      
.P     8 ` + "`" + `YooP' 8    8 ` + "`" + `Yooo' ` + "`" + `YooP'   ` + "`" + `YooP'  8       8   8    8 ` + "`" + `Yooo'  
..:::::..:.....:..:::..:.....::.....::::.....::..::::::..::..:::..:.....::
::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::
::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::`

	logoPart2 := `      ooooo        8 8                .oPYo.             8                     
      8            8 8                8    8             8                     
      o8oo   .oPYo. 8 8 .oPYo. odYo.   8      .oPYo. .oPYo8 .oPYo.            
      8     .oooo8 8 8 8oooo8 8' ` + "`" + `8   8   oo 8    8 8    8 Yb..               
      8     8    8 8 8 8.     8   8   8    8 8    8 8    8   'Yb.             
      8     ` + "`" + `YooP8 8 8 ` + "`" + `Yooo' 8   8   ` + "`" + `YooP8 ` + "`" + `YooP' ` + "`" + `YooP' ` + "`" + `YooP'             
      :..:::::.....:....:.....:..::..:::....8 :.....::.....::.....:            
      ::::::::::::::::::::::::::::::::::::::8 :::::::::::::::::::::            
      ::::::::::::::::::::::::::::::::::::::..:::::::::::::::::::::            `

	fmt.Fprint(v, centerText(logoPart1, maxX)+"\n"+centerText(logoPart2, maxX))

	if !inForm {
		menuY := 20
		menuHeight := len(menuItems) + 2
		menuWidth := 30
		menuX := (maxX - menuWidth) / 2

		if v, err := g.SetView("menu", menuX, menuY, menuX+menuWidth, menuY+menuHeight, 0); err != nil {
			if !errors.Is(err, gocui.ErrUnknownView) {
				return err
			}
			v.Frame = true
			v.Title = " Main Menu "
			v.Highlight = false

			for i, item := range menuItems {
				if i == selected {
					fmt.Fprintf(v, "\033[7m%s\033[0m\n", item.text)
				} else {
					fmt.Fprintln(v, item.text)
				}
			}

			g.SetCurrentView("menu")
		} else {
			v.Clear()
			for i, item := range menuItems {
				if i == selected {
					fmt.Fprintf(v, "\033[7m%s\033[0m\n", item.text)
				} else {
					fmt.Fprintln(v, item.text)
				}
			}
		}

		if v, err := g.SetView("instructions", 0, maxY-3, maxX-1, maxY-1, 0); err != nil {
			if !errors.Is(err, gocui.ErrUnknownView) {
				return err
			}
			v.Frame = false
			fmt.Fprintln(v, "Use arrows/mouse to navigate, Enter to select, Ctrl+C to quit")
		}
	} else {
		formWidth := 60
		formHeight := 18
		formX := (maxX - formWidth) / 2
		formY := (maxY - formHeight) / 2

		if v, err := g.SetView("form", formX, formY, formX+formWidth, formY+formHeight, 0); err != nil {
			if !errors.Is(err, gocui.ErrUnknownView) {
				return err
			}
			v.Frame = true
			v.Title = " New Game "
		}

		labelY := formY + 2
		createFormField(g, "username_label", "username", "Username:", formX+2, labelY, formX+formWidth-2, labelY+2, true)

		labelY += 3
		createFormField(g, "race_label", "race", "Race:", formX+2, labelY, formX+formWidth-2, labelY+2, false)

		labelY += 3
		createFormField(g, "seed_label", "seed", "Seed:", formX+2, labelY, formX+formWidth-2, labelY+2, true)

		buttonY := formY + formHeight - 5
		createButton(g, "save_button", " Save ", formX+15, buttonY-1, 10, 2, "save_button")
		createButton(g, "cancel_button", " Cancel ", formX+35, buttonY-1, 10, 2, "cancel_button")

		errorY := formY + formHeight - 4
		if v, err := g.SetView("error_message", formX+2, errorY, formX+formWidth-2, errorY+2, 0); err != nil {
			if !errors.Is(err, gocui.ErrUnknownView) {
				return err
			}
			v.Frame = false
			if errMsg != "" {
				fmt.Fprintf(v, "\033[31m%s\033[0m", errMsg)
			}
		} else {
			v.Clear()
			if errMsg != "" {
				fmt.Fprintf(v, "\033[31m%s\033[0m", errMsg)
			}
		}

		instructionY := formY + formHeight - 2
		if v, err := g.SetView("form_instructions", formX+2, instructionY, formX+formWidth-2, instructionY+2, 0); err != nil {
			if !errors.Is(err, gocui.ErrUnknownView) {
				return err
			}
			v.Frame = false
			fmt.Fprint(v, "← → arrows to select race")
		}

		if fieldIndex < len(fields) {
			g.SetCurrentView(fields[fieldIndex])
		}
	}

	return nil
}

func showNewGameForm(g *gocui.Gui, v *gocui.View) error {
	inForm = true
	fieldIndex = 0
	username = ""
	seed = ""
	raceIndex = 0
	errMsg = ""
	g.DeleteView("menu")
	g.DeleteView("instructions")
	return nil
}

func showDialog(g *gocui.Gui, title, message string) error {
	dialog = true
	maxX, maxY := g.Size()
	msgWidth := 50
	msgHeight := 10
	msgX := (maxX - msgWidth) / 2
	msgY := (maxY - msgHeight) / 2

	if v, err := g.SetView("message", msgX, msgY, msgX+msgWidth, msgY+msgHeight, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Frame = true
		v.Title = title
		fmt.Fprintln(v, "\n  "+message)
		fmt.Fprintln(v, "  (Not implemented)")
		fmt.Fprintln(v, "")

		g.SetKeybinding("message", gocui.KeyEsc, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
			dialog = false
			return g.DeleteView("message")
		})
	}

	buttonY := msgY + msgHeight - 2
	buttonX := msgX + msgWidth - 14
	if v, err := g.SetView("go_back_button", buttonX, buttonY-1, buttonX+12, buttonY+1, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Frame = true
		if hovered == "go_back_button" {
			v.BgColor = gocui.ColorYellow
			v.FgColor = gocui.ColorBlack
		}
		fmt.Fprint(v, " Go Back ")
	}

	g.SetKeybinding("go_back_button", gocui.KeyEnter, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		dialog = false
		g.DeleteView("message")
		g.DeleteView("go_back_button")
		return nil
	})

	return nil
}

func createButton(g *gocui.Gui, name, text string, x, y, w, h int, hoverName string) error {
	if v, err := g.SetView(name, x, y, x+w, y+h, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Frame = true
		if hovered == hoverName {
			v.BgColor = gocui.ColorYellow
			v.FgColor = gocui.ColorBlack
		}
		fmt.Fprint(v, text)
	}
	return nil
}

func createFormField(g *gocui.Gui, labelName, fieldName, labelText string, x, y, w, h int, editable bool) error {
	// Create label
	if v, err := g.SetView(labelName, x, y, x+10, y+2, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Frame = false
		fmt.Fprint(v, labelText)
	}

	// Create field with appropriate width (30 characters max)
	fieldWidth := 50
	if fieldWidth > w-x-11 {
		fieldWidth = w - x - 11
	}
	if v, err := g.SetView(fieldName, x+11, y, x+11+fieldWidth, y+2, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Frame = true
		v.Editable = editable
		v.Wrap = false
		if fieldName == "username" {
			fmt.Fprint(v, username)
		} else if fieldName == "seed" {
			fmt.Fprint(v, seed)
		} else if fieldName == "race" {
			fmt.Fprintf(v, "< %s >", races[raceIndex])
		}
	} else if fieldName == "race" {
		v.Clear()
		fmt.Fprintf(v, "< %s >", races[raceIndex])
	}
	return nil
}

func loadGameFile(g *gocui.Gui, v *gocui.View) error {
	return showDialog(g, " Load Game ", "Loading game file...")
}

func eraseGameFile(g *gocui.Gui, v *gocui.View) error {
	return showDialog(g, " Erase File ", "Erasing game file...")
}

func saveForm(g *gocui.Gui) error {
	checkInput(g)
	if errMsg != "" {
		return nil
	}

	v, _ := g.View("username")
	username = v.Buffer()
	v, _ = g.View("seed")
	seed = v.Buffer()
	race = races[raceIndex]

	maxX, maxY := g.Size()
	msgWidth := 50
	msgHeight := 10
	msgX := (maxX - msgWidth) / 2
	msgY := (maxY - msgHeight) / 2

	if v, err := g.SetView("saved", msgX, msgY, msgX+msgWidth, msgY+msgHeight, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Frame = true
		v.Title = " Game Saved "
		fmt.Fprintln(v, "\n[Game Data Saved]")
		fmt.Fprintf(v, "Username: %s\n", username)
		fmt.Fprintf(v, "Race: %s\n", race)
		fmt.Fprintf(v, "Seed: %s\n", seed)
		fmt.Fprintln(v, "")
	}

	buttonY := msgY + msgHeight - 2
	buttonX := msgX + msgWidth - 14
	createButton(g, "saved_go_back_button", " Go Back ", buttonX, buttonY-1, 12, 2, "go_back_button")

	return nil
}

func cancelForm(g *gocui.Gui) error {
	inForm = false
	errMsg = ""
	// cleanup form views
	views := []string{"form", "username_label", "username", "race_label", "race", "seed_label", "seed", "save_button", "cancel_button", "form_instructions", "error_message", "saved", "saved_go_back_button"}
	for _, v := range views {
		g.DeleteView(v)
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func keybindings(g *gocui.Gui) error {
	g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit)

	g.SetKeybinding("menu", gocui.KeyArrowUp, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if selected > 0 {
			selected--
		}
		return nil
	})

	g.SetKeybinding("menu", gocui.KeyArrowDown, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if selected < len(menuItems)-1 {
			selected++
		}
		return nil
	})

	g.SetKeybinding("menu", gocui.KeyEnter, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return menuItems[selected].action(g, v)
	})

	g.SetKeybinding("race", gocui.KeyArrowLeft, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if raceIndex > 0 {
			raceIndex--
		}
		raceView, _ := g.View("race")
		raceView.Clear()
		fmt.Fprintf(raceView, "< %s >", races[raceIndex])
		return nil
	})

	g.SetKeybinding("race", gocui.KeyArrowRight, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if raceIndex < len(races)-1 {
			raceIndex++
		}
		raceView, _ := g.View("race")
		raceView.Clear()
		fmt.Fprintf(raceView, "< %s >", races[raceIndex])
		return nil
	})

	g.SetKeybinding("username", gocui.KeyEnter, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		checkInput(g)
		return nil
	})

	g.SetKeybinding("seed", gocui.KeyEnter, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		checkInput(g)
		return nil
	})

	g.SetKeybinding("", gocui.KeyEsc, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if inForm {
			return cancelForm(g)
		}
		return nil
	})

	g.SetKeybinding("saved_go_back_button", gocui.KeyEnter, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		g.DeleteView("saved")
		g.DeleteView("saved_go_back_button")
		return cancelForm(g)
	})

	g.SetKeybinding("saved", gocui.KeyEsc, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		g.DeleteView("saved")
		g.DeleteView("saved_go_back_button")
		return cancelForm(g)
	})

	g.Mouse = true
	g.SetKeybinding("", gocui.MouseLeft, gocui.ModNone, handleMouseClick)
	return nil
}

func isMouseOver(g *gocui.Gui, viewName string, mx, my int) bool {
	view, _ := g.View(viewName)
	if view == nil {
		return false
	}
	x0, y0, x1, y1 := view.Dimensions()
	return mx >= x0 && mx <= x1 && my >= y0 && my <= y1
}

func updateHoverEffects(g *gocui.Gui) {
	mx, my := g.MousePosition()
	needsUpdate := false

	if !inForm && !dialog {
		menuView, _ := g.View("menu")
		if menuView != nil {
			x0, y0, x1, y1 := menuView.Dimensions()
			if mx >= x0 && mx <= x1 && my >= y0 && my <= y1 {
				itemIndex := my - y0 - 1
				if itemIndex >= 0 && itemIndex < len(menuItems) && itemIndex != selected {
					selected = itemIndex
					needsUpdate = true
				}
			}
		}
	} else {
		newHoveredButton := ""
		buttonNames := []string{"save_button", "cancel_button", "saved_go_back_button", "go_back_button"}

		for _, btnName := range buttonNames {
			if isMouseOver(g, btnName, mx, my) {
				newHoveredButton = btnName
				break
			}
		}

		if newHoveredButton != hovered {
			hovered = newHoveredButton
			needsUpdate = true
		}
	}

	if needsUpdate {
		g.Update(func(g *gocui.Gui) error {
			return nil
		})
	}
}

func handleMouseClick(g *gocui.Gui, v *gocui.View) error {
	mx, my := g.MousePosition()
	updateHoverEffects(g)

	if !inForm && !dialog {
		menuView, _ := g.View("menu")
		x0, y0, x1, y1 := menuView.Dimensions()
		if mx >= x0 && mx <= x1 && my >= y0 && my <= y1 {
			itemIndex := my - y0 - 1
			if itemIndex >= 0 && itemIndex < len(menuItems) {
				if itemIndex != selected {
					selected = itemIndex
				}
				return menuItems[selected].action(g, v)
			}
		}
	} else {
		views := []string{"username", "race", "seed"}
		for _, viewName := range views {
			view, _ := g.View(viewName)
			if view != nil {
				x0, y0, x1, y1 := view.Dimensions()
				if mx >= x0 && mx <= x1 && my >= y0 && my <= y1 {
					for i, field := range fields {
						if field == viewName {
							fieldIndex = i
							g.SetCurrentView(viewName)
							break
						}
					}
				}
			}
		}

		if isMouseOver(g, "save_button", mx, my) {
			hovered = ""
			return saveForm(g)
		}

		if isMouseOver(g, "cancel_button", mx, my) {
			hovered = ""
			return cancelForm(g)
		}

		if isMouseOver(g, "saved_go_back_button", mx, my) {
			hovered = ""
			g.DeleteView("saved")
			g.DeleteView("saved_go_back_button")
			return cancelForm(g)
		}
	}

	if isMouseOver(g, "go_back_button", mx, my) {
		hovered = ""
		dialog = false
		g.DeleteView("message")
		g.DeleteView("go_back_button")
		return nil
	}

	return nil
}

func main() {
	setTerminalSize(150, 38)
	g, _ = gocui.NewGui(gocui.OutputNormal, false)
	defer g.Close()
	g.SetManagerFunc(layout)
	keybindings(g)
	g.MainLoop()
}
