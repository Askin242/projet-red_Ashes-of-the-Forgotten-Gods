package ui

import (
	"errors"
	"fmt"
	"main/pkg/save"
	"main/pkg/structures"
	"math/rand"
	"os"
	"time"

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

	if getLengthFromString(user) > 10 {
		errMsg = "Username is too long, max 10 characters"
		return
	}

	if getLengthFromString(seed) > 10 {
		errMsg = "Seed is too long, max 10 characters"
		return
	}
}

func generateRandomSeed() string {
	tempRNG := rand.New(rand.NewSource(time.Now().UnixNano()))
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	result := make([]byte, 10)
	for i := range result {
		result[i] = charset[tempRNG.Intn(len(charset))]
	}
	return string(result)
}

func SaveGameState() error {
	if !structures.IsSeeded() {
		return nil
	}

	structures.RefreshSeedState()
	gameConfig := save.GameConfig{
		Username:    username,
		Race:        race,
		Seed:        seed,
		CurrentSeed: structures.GetCurrentSeedState(),
	}

	return save.SaveGameConfig(gameConfig)
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
		CreateFormField(g, "username_label", "username", "Username:", formX+2, labelY, formX+formWidth-2, labelY+2, true, username, seed, races, raceIndex)

		labelY += 3
		CreateFormField(g, "race_label", "race", "Race:", formX+2, labelY, formX+formWidth-2, labelY+2, false, username, seed, races, raceIndex)

		labelY += 3
		if v, err := g.SetView("seed_label", formX+2, labelY, formX+12, labelY+2, 0); err != nil {
			if !errors.Is(err, gocui.ErrUnknownView) {
				return err
			}
			v.Frame = false
			fmt.Fprint(v, "Seed:")
		}

		if v, err := g.SetView("seed", formX+13, labelY, formX+33, labelY+2, 0); err != nil {
			if !errors.Is(err, gocui.ErrUnknownView) {
				return err
			}
			v.Frame = true
			v.Editable = true
			v.Wrap = false
			fmt.Fprint(v, seed)
		}

		createButton(g, "generate_seed_button", " Random ", formX+35, labelY, 10, 2, "generate_seed_button")

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
	return ShowSimpleDialog(g, "main", title, message, " Go Back ", 50, 10, func() {
		dialog = false
	})
}

var (
	loadSaves    []save.SaveInfo
	selectedSave int
)

func showLoadDialog(g *gocui.Gui) error {
	dialog = true
	selectedSave = 0

	saves, err := save.GetAvailableSaves()
	if err != nil {
		return showDialog(g, " Error ", "Failed to load saves: "+err.Error())
	}

	loadSaves = saves

	if len(loadSaves) == 0 {
		return showDialog(g, " No Saves ", "No saved games found.")
	}

	maxX, maxY := g.Size()
	dialogWidth := 60
	dialogHeight := min(18, len(loadSaves)+10)
	dialogX := (maxX - dialogWidth) / 2
	dialogY := (maxY - dialogHeight) / 2

	if v, err := g.SetView("load_dialog", dialogX, dialogY, dialogX+dialogWidth, dialogY+dialogHeight, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Frame = true
		v.Title = " Load Game "
		fmt.Fprintln(v, "\nSelect a save to load:")
		fmt.Fprintln(v, "")

		for i, saveInfo := range loadSaves {
			if i == selectedSave {
				fmt.Fprintf(v, "\033[7m[%d] %s (%s) - Seed: %s\033[0m\n",
					i+1, saveInfo.Username, saveInfo.Race, saveInfo.Seed)
			} else {
				fmt.Fprintf(v, "[%d] %s (%s) - Seed: %s\n",
					i+1, saveInfo.Username, saveInfo.Race, saveInfo.Seed)
			}
		}

		fmt.Fprintln(v, "")
		fmt.Fprintln(v, "Use ↑↓ arrows to select")
	}

	buttonY := dialogY + dialogHeight - 4
	createButton(g, "load_confirm_button", " Load ", dialogX+15, buttonY, 10, 2, "load_confirm_button")
	createButton(g, "load_cancel_button", " Cancel ", dialogX+35, buttonY, 10, 2, "load_cancel_button")

	g.SetKeybinding("load_dialog", gocui.KeyArrowUp, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if selectedSave > 0 {
			selectedSave--
			refreshLoadDialog(g)
		}
		return nil
	})

	g.SetKeybinding("load_dialog", gocui.KeyArrowDown, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if selectedSave < len(loadSaves)-1 {
			selectedSave++
			refreshLoadDialog(g)
		}
		return nil
	})

	g.SetKeybinding("load_dialog", gocui.KeyEnter, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return confirmLoad(g)
	})

	g.SetKeybinding("load_confirm_button", gocui.KeyEnter, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return confirmLoad(g)
	})

	g.SetKeybinding("load_cancel_button", gocui.KeyEnter, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return cancelLoad(g)
	})

	g.SetCurrentView("load_dialog")
	return nil
}

func refreshLoadDialog(g *gocui.Gui) {
	v, _ := g.View("load_dialog")
	if v != nil {
		v.Clear()
		fmt.Fprintln(v, "\nSelect a save to load:")
		fmt.Fprintln(v, "")

		for i, saveInfo := range loadSaves {
			if i == selectedSave {
				fmt.Fprintf(v, "\033[7m[%d] %s (%s) - Seed: %s\033[0m\n",
					i+1, saveInfo.Username, saveInfo.Race, saveInfo.Seed)
			} else {
				fmt.Fprintf(v, "[%d] %s (%s) - Seed: %s\n",
					i+1, saveInfo.Username, saveInfo.Race, saveInfo.Seed)
			}
		}

		fmt.Fprintln(v, "")
		fmt.Fprintln(v, "Use ↑↓ arrows to select")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func confirmLoad(g *gocui.Gui) error {
	if selectedSave < 0 || selectedSave >= len(loadSaves) {
		return nil
	}

	selectedSaveInfo := loadSaves[selectedSave]

	save.SetSaveID(selectedSaveInfo.Username)
	config, err := save.LoadGameConfig()
	if err != nil {
		return showDialog(g, " Error ", "Failed to load save: "+err.Error())
	}

	username = config.Username
	race = config.Race
	seed = config.Seed

	if seed != "" {
		if config.CurrentSeed != 0 {
			// Load from saved RNG state
			structures.InitializeFromCurrentSeed(seed, config.CurrentSeed)
		} else {
			// Fallback for old saves without CurrentSeed
			structures.InitializeSeed(seed)
		}
	}

	for i, r := range races {
		if r == race {
			raceIndex = i
			break
		}
	}

	return showLoadSuccess(g, config)
}

func showLoadSuccess(g *gocui.Gui, config save.GameConfig) error {
	views := []string{"load_dialog", "load_confirm_button", "load_cancel_button"}
	for _, viewName := range views {
		g.DeleteView(viewName)
	}

	maxX, maxY := g.Size()
	msgWidth := 50
	msgHeight := 12
	msgX := (maxX - msgWidth) / 2
	msgY := (maxY - msgHeight) / 2

	if v, err := g.SetView("loaded", msgX, msgY, msgX+msgWidth, msgY+msgHeight, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Frame = true
		v.Title = " Game Loaded "
		fmt.Fprintln(v, "\n[Game Loaded Successfully]")
		fmt.Fprintf(v, "Username: %s\n", config.Username)
		fmt.Fprintf(v, "Race: %s\n", config.Race)
		fmt.Fprintf(v, "Seed: %s\n", config.Seed)
		fmt.Fprintf(v, "Loaded from: saves/%s/\n", config.Username)
		fmt.Fprintln(v, "")
	}

	buttonY := msgY + msgHeight - 2
	buttonX := msgX + msgWidth - 14
	createButton(g, "loaded_go_back_button", " Go Back ", buttonX, buttonY-1, 12, 2, "loaded_go_back_button")

	g.SetKeybinding("loaded_go_back_button", gocui.KeyEnter, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		g.DeleteView("loaded")
		g.DeleteView("loaded_go_back_button")
		dialog = false
		return nil
	})

	return nil
}

func cancelLoad(g *gocui.Gui) error {
	views := []string{"load_dialog", "load_confirm_button", "load_cancel_button"}
	for _, viewName := range views {
		g.DeleteView(viewName)
	}
	dialog = false
	return nil
}

func loadGameFile(g *gocui.Gui, v *gocui.View) error {
	return showLoadDialog(g)
}

func eraseGameFile(g *gocui.Gui, v *gocui.View) error {
	return showEraseDialog(g)
}

var selectedErase int

func showEraseDialog(g *gocui.Gui) error {
	dialog = true
	selectedErase = 0

	saves, err := save.GetAvailableSaves()
	if err != nil {
		return showDialog(g, " Error ", "Failed to load saves: "+err.Error())
	}

	if len(saves) == 0 {
		return showDialog(g, " No Saves ", "No saved games to erase.")
	}

	loadSaves = saves
	maxX, maxY := g.Size()
	dialogWidth := 60
	dialogHeight := min(18, len(saves)+10)
	dialogX := (maxX - dialogWidth) / 2
	dialogY := (maxY - dialogHeight) / 2

	if v, err := g.SetView("erase_dialog", dialogX, dialogY, dialogX+dialogWidth, dialogY+dialogHeight, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Frame = true
		v.Title = " Erase Game "
		refreshEraseDialog(g)
	}

	buttonY := dialogY + dialogHeight - 3
	createButton(g, "erase_confirm_button", " ERASE ", dialogX+15, buttonY, 10, 2, "erase_confirm_button")
	createButton(g, "erase_cancel_button", " Cancel ", dialogX+35, buttonY, 10, 2, "erase_cancel_button")

	g.SetKeybinding("erase_dialog", gocui.KeyArrowUp, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if selectedErase > 0 {
			selectedErase--
			refreshEraseDialog(g)
		}
		return nil
	})

	g.SetKeybinding("erase_dialog", gocui.KeyArrowDown, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if selectedErase < len(loadSaves)-1 {
			selectedErase++
			refreshEraseDialog(g)
		}
		return nil
	})

	g.SetKeybinding("erase_confirm_button", gocui.KeyEnter, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return confirmErase(g)
	})

	g.SetKeybinding("erase_cancel_button", gocui.KeyEnter, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return cancelErase(g)
	})

	g.SetKeybinding("erase_dialog", gocui.KeyEnter, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return confirmErase(g)
	})

	g.SetCurrentView("erase_dialog")
	return nil
}

func refreshEraseDialog(g *gocui.Gui) {
	v, _ := g.View("erase_dialog")
	if v != nil {
		v.Clear()
		fmt.Fprintln(v, "\n\033[31mWARNING: This will permanently delete the save!\033[0m")
		fmt.Fprintln(v, "Select a save to erase:")
		fmt.Fprintln(v, "")

		for i, saveInfo := range loadSaves {
			if i == selectedErase {
				fmt.Fprintf(v, "\033[7m[%d] %s (%s) - Seed: %s\033[0m\n",
					i+1, saveInfo.Username, saveInfo.Race, saveInfo.Seed)
			} else {
				fmt.Fprintf(v, "[%d] %s (%s) - Seed: %s\n",
					i+1, saveInfo.Username, saveInfo.Race, saveInfo.Seed)
			}
		}

		fmt.Fprintln(v, "")
		fmt.Fprintln(v, "Use ↑↓ arrows to select")
	}
}

func confirmErase(g *gocui.Gui) error {
	if selectedErase < 0 || selectedErase >= len(loadSaves) {
		return nil
	}

	selectedSaveInfo := loadSaves[selectedErase]

	savePath := "saves/" + selectedSaveInfo.Username
	err := os.RemoveAll(savePath)
	if err != nil {
		return showDialog(g, " Error ", "Failed to erase save: "+err.Error())
	}

	views := []string{"erase_dialog", "erase_confirm_button", "erase_cancel_button"}
	for _, viewName := range views {
		g.DeleteView(viewName)
	}

	return showDialog(g, " Success ", "Save '"+selectedSaveInfo.Username+"' has been erased.")
}

func cancelErase(g *gocui.Gui) error {
	views := []string{"erase_dialog", "erase_confirm_button", "erase_cancel_button"}
	for _, viewName := range views {
		g.DeleteView(viewName)
	}
	dialog = false
	return nil
}

func saveForm(g *gocui.Gui) error {
	checkInput(g)
	if errMsg != "" {
		return nil
	}

	v, _ := g.View("username")
	usernameToCheck := v.Buffer()
	if save.SaveExists(usernameToCheck) {
		errMsg = "Username already exists! Choose a different one."
		return nil
	}

	v, _ = g.View("username")
	username = v.Buffer()
	v, _ = g.View("seed")
	seed = v.Buffer()
	race = races[raceIndex]

	if seed == "" {
		seed = generateRandomSeed()
	}
	structures.InitializeSeed(seed)

	save.SetSaveID(username)
	gameConfig := save.GameConfig{
		Username:    username,
		Race:        race,
		Seed:        seed,
		CurrentSeed: structures.GetCurrentSeedState(),
	}

	err := save.SaveGameConfig(gameConfig)
	if err != nil {
		errMsg = "Failed to save game: " + err.Error()
		return nil
	}

	maxX, maxY := g.Size()
	msgWidth := 50
	msgHeight := 12
	msgX := (maxX - msgWidth) / 2
	msgY := (maxY - msgHeight) / 2

	if v, err := g.SetView("saved", msgX, msgY, msgX+msgWidth, msgY+msgHeight, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Frame = true
		v.Title = " Game Saved "
		fmt.Fprintln(v, "\n[Game Data Saved Successfully]")
		fmt.Fprintf(v, "Username: %s\n", username)
		fmt.Fprintf(v, "Race: %s\n", race)
		fmt.Fprintf(v, "Seed: %s\n", seed)
		fmt.Fprintf(v, "Saved to: saves/%s/\n", username)
		fmt.Fprintln(v, "")
	}

	buttonY := msgY + msgHeight - 2
	buttonX := msgX + msgWidth - 14
	createButton(g, "saved_go_back_button", " Go Back ", buttonX, buttonY-1, 12, 2, "go_back_button")

	return nil
}

func cancelForm(g *gocui.Gui) error {
	inForm = false
	dialog = false
	errMsg = ""
	views := []string{"form", "username_label", "username", "race_label", "race", "seed_label", "seed", "generate_seed_button", "save_button", "cancel_button", "form_instructions", "error_message", "saved", "saved_go_back_button", "load_dialog", "load_confirm_button", "load_cancel_button", "loaded", "loaded_go_back_button", "erase_dialog", "erase_confirm_button", "erase_cancel_button"}
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

	g.SetKeybinding("generate_seed_button", gocui.KeyEnter, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		randomSeed := generateRandomSeed()
		seedView, _ := g.View("seed")
		if seedView != nil {
			seedView.Clear()
			fmt.Fprint(seedView, randomSeed)
		}
		return nil
	})

	EnableMouseAndSetHandler(g, handleMouseClick)
	return nil
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
		buttonNames := []string{"generate_seed_button", "save_button", "cancel_button", "saved_go_back_button", "main_dialog_btn", "load_confirm_button", "load_cancel_button", "loaded_go_back_button", "erase_confirm_button", "erase_cancel_button"}

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

		buttons := []ButtonHandlerWithHover{
			{"generate_seed_button", func(g *gocui.Gui, v *gocui.View) error {
				randomSeed := generateRandomSeed()
				seedView, _ := g.View("seed")
				if seedView != nil {
					seedView.Clear()
					fmt.Fprint(seedView, randomSeed)
				}
				return nil
			}, func() { hovered = "" }},
			{"save_button", func(g *gocui.Gui, v *gocui.View) error { return saveForm(g) }, func() { hovered = "" }},
			{"cancel_button", func(g *gocui.Gui, v *gocui.View) error { return cancelForm(g) }, func() { hovered = "" }},
			{"saved_go_back_button", func(g *gocui.Gui, v *gocui.View) error {
				g.DeleteView("saved")
				g.DeleteView("saved_go_back_button")
				return cancelForm(g)
			}, func() { hovered = "" }},
			{"load_confirm_button", func(g *gocui.Gui, v *gocui.View) error { return confirmLoad(g) }, func() { hovered = "" }},
			{"load_cancel_button", func(g *gocui.Gui, v *gocui.View) error { return cancelLoad(g) }, func() { hovered = "" }},
			{"loaded_go_back_button", func(g *gocui.Gui, v *gocui.View) error {
				g.DeleteView("loaded")
				g.DeleteView("loaded_go_back_button")
				dialog = false
				return nil
			}, func() { hovered = "" }},
			{"erase_confirm_button", func(g *gocui.Gui, v *gocui.View) error { return confirmErase(g) }, func() { hovered = "" }},
			{"erase_cancel_button", func(g *gocui.Gui, v *gocui.View) error { return cancelErase(g) }, func() { hovered = "" }},
		}
		if err := HandleMouseClickButtonsWithHover(g, mx, my, buttons); err != nil {
			return err
		}

		loadView, _ := g.View("load_dialog")
		if loadView != nil && len(loadSaves) > 0 {
			x0, y0, x1, y1 := loadView.Dimensions()
			if mx >= x0 && mx <= x1 && my >= y0 && my <= y1 {
				saveLineStart := y0 + 3
				if my >= saveLineStart && my < saveLineStart+len(loadSaves) {
					clickedSave := my - saveLineStart
					if clickedSave >= 0 && clickedSave < len(loadSaves) {
						selectedSave = clickedSave
						refreshLoadDialog(g)
					}
				}
			}
		}

		eraseView, _ := g.View("erase_dialog")
		if eraseView != nil && len(loadSaves) > 0 {
			x0, y0, x1, y1 := eraseView.Dimensions()
			if mx >= x0 && mx <= x1 && my >= y0 && my <= y1 {
				saveLineStart := y0 + 4
				if my >= saveLineStart && my < saveLineStart+len(loadSaves) {
					clickedSave := my - saveLineStart
					if clickedSave >= 0 && clickedSave < len(loadSaves) {
						selectedErase = clickedSave
						refreshEraseDialog(g)
					}
				}
			}
		}
	}

	mainButtons := []ButtonHandlerWithHover{
		{"main_dialog_btn", func(g *gocui.Gui, v *gocui.View) error {
			dialog = false
			g.DeleteView("main_dialog")
			g.DeleteView("main_dialog_btn")
			return nil
		}, func() { hovered = "" }},
	}
	if err := HandleMouseClickButtonsWithHover(g, mx, my, mainButtons); err != nil {
		return err
	}

	return nil
}

func ShowMainMenu() {
	g, _ = gocui.NewGui(gocui.OutputNormal, false)
	defer g.Close()
	g.SetManagerFunc(layout)
	keybindings(g)
	g.MainLoop()
}
