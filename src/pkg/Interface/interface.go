package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	username   string
	chosenRace string
	customSeed string
)

func setTerminalSize(cols, rows int) {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/C", fmt.Sprintf("mode con: cols=%d lines=%d", cols, rows))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		_ = cmd.Run()
	default:
		fmt.Printf("\033[8;%d;%dt", rows, cols)
	}
}

func printCentered(lines []string, cols int, extraRight int) {
	maxLength := 0
	for _, line := range lines {
		trimmed := strings.TrimLeft(line, " ")
		if len(trimmed) > maxLength {
			maxLength = len(trimmed)
		}
	}

	consolePadding := 0
	if cols > maxLength {
		consolePadding = (cols - maxLength) / 2
	}

	for _, line := range lines {
		trimmed := strings.TrimLeft(line, " ")
		padding := (maxLength - len(trimmed)) / 2
		fmt.Printf("%s%s%s\n", strings.Repeat(" ", consolePadding+padding+extraRight), trimmed, strings.Repeat(" ", padding))
	}
}

func printText() {
	cols := 160
	logo := `                                                                           
     .oo        8                      .oPYo.  d'b   ooooo 8               
    .P 8        8                      8    8  8       8   8               
   .P  8 .oPYo. 8oPYo. .oPYo. .oPYo.   8    8 o8P      8   8oPYo. .oPYo.  
  oPooo8 Yb..   8    8 8oooo8 Yb..     8    8  8       8   8    8 8oooo8  
 .P    8   'Yb. 8    8 8.       'Yb.   8    8  8       8   8    8 8.      
.P     8 ` + "`" + `YooP' 8    8 ` + "`" + `Yooo' ` + "`" + `YooP'   ` + "`" + `YooP'  8       8   8    8 ` + "`" + `Yooo'  
..:::::..:.....:..:::..:.....::.....::::.....::..::::::..::..:::..:.....::
::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::
::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::
                                                                            `

	fallen := []string{
		"ooooo        8 8                .oPYo.             8                     ",
		"8            8 8                8    8             8                     ",
		"o8oo   .oPYo. 8 8 .oPYo. odYo.   8      .oPYo. .oPYo8 .oPYo.            ",
		"8     .oooo8 8 8 8oooo8 8' `8   8   oo 8    8 8    8 Yb..               ",
		"8     8    8 8 8 8.     8   8   8    8 8    8 8    8   'Yb.             ",
		"8     `YooP8 8 8 `Yooo' 8   8   `YooP8 `YooP' `YooP' `YooP'             ",
		":..:::::.....:....:.....:..::..:::....8 :.....::.....::.....:            ",
		"::::::::::::::::::::::::::::::::::::::8 :::::::::::::::::::::            ",
		"::::::::::::::::::::::::::::::::::::::..:::::::::::::::::::::            ",
	}

	printCentered(strings.Split(logo, "\n"), cols, 0)
	fmt.Println()
	printCentered(fallen, cols, 6)
}

func newGameForm() {
	app := tview.NewApplication()
	form := tview.NewForm()

	// Customize styles (dark background with accent)
	form.SetBackgroundColor(tcell.ColorBlack)                 // overall form background
	form.SetLabelColor(tcell.ColorLightGray)                  // label text
	form.SetFieldBackgroundColor(tcell.ColorGray.TrueColor()) // input box background
	form.SetFieldTextColor(tcell.ColorWhite)                  // input text color
	form.SetButtonBackgroundColor(tcell.ColorDarkSlateGray)
	form.SetButtonTextColor(tcell.ColorAqua)

	form.AddInputField("Username:", "", 25, nil, func(text string) {
		username = text
	})

	form.AddDropDown("Race:", []string{"Human", "Elf", "Dwarf"}, 0, func(option string, index int) {
		chosenRace = option
	})

	form.AddInputField("Custom seed:", "", 10, nil, func(text string) {
		customSeed = text
	})

	form.AddButton("Save", func() {
		fmt.Println("\n[Saved Game Data]")
		fmt.Println("Username:", username)
		fmt.Println("Race:", chosenRace)
		fmt.Println("Seed:", customSeed)
		app.Stop()
	})

	app.SetRoot(form, true)
	if err := app.Run(); err != nil {
		panic(err)
	}
}

func printOptions() {
	fmt.Println()
	fmt.Println("[1] New game file")
	fmt.Println("[2] Load game file")
	fmt.Println("[3] Erase game file")
	fmt.Println("[4] Exit")

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Select an option> ")
	mode, _ := reader.ReadString('\n')
	mode = strings.TrimSpace(mode)

	switch mode {
	case "1":
		newGameForm()
	case "2":
		fmt.Println("load game file")
	case "3":
		fmt.Println("erase file")
	case "4":
		fmt.Println("exit")
	default:
		fmt.Println("Invalid option")
	}
}

func main() {
	setTerminalSize(150, 38)
	printText()
	printOptions()
}
