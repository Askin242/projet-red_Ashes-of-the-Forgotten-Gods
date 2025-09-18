package ui

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"errors"

	"github.com/awesome-gocui/gocui"
	"golang.org/x/sys/windows"
)

func enableAnsiColors() {
	if runtime.GOOS == "windows" {
		var originalMode uint32
		stdout := windows.Handle(os.Stdout.Fd())
		windows.GetConsoleMode(stdout, &originalMode)
		windows.SetConsoleMode(stdout, originalMode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
	}
}

func setTerminalSize(cols, rows int) {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/C", fmt.Sprintf("mode con: cols=%d lines=%d", cols, rows))
		cmd.Run()
	} else {
		fmt.Printf("\033[8;%d;%dt", rows, cols)
	}
}

func ClearScreen() {
	switch runtime.GOOS {
	case "linux":
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func InitScreen() {
	setTerminalSize(150, 38)
	enableAnsiColors()
	ClearScreen()
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

func getLengthFromString(s string) int {
	counter := 0
	for range s {
		counter++
	}
	return counter
}

func RenderListWithHighlight(v *gocui.View, lines []string, selected int) {
	v.Clear()
	for i, line := range lines {
		if i == selected {
			fmt.Fprintf(v, "\u001b[7m%s\u001b[0m\n", line)
		} else {
			fmt.Fprintln(v, line)
		}
	}
}

func BindListNavigation(g *gocui.Gui, viewName string, selected *int, count func() int) error {
	if err := g.SetKeybinding(viewName, gocui.KeyArrowUp, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if *selected > 0 {
			*selected--
		}
		return nil
	}); err != nil {
		return err
	}
	if err := g.SetKeybinding(viewName, gocui.KeyArrowDown, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		n := count()
		if n <= 0 {
			return nil
		}
		if *selected < n-1 {
			*selected++
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func SetOrUpdateView(g *gocui.Gui, name string, x0, y0, x1, y1 int, init func(v *gocui.View), update func(v *gocui.View)) error {
	if v, err := g.SetView(name, x0, y0, x1, y1, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		if init != nil {
			init(v)
		}
		if update != nil {
			update(v)
		}
	} else {
		v.Clear()
		if update != nil {
			update(v)
		}
	}
	return nil
}

func BindQuitOnEsc(g *gocui.Gui) error {
	return g.SetKeybinding("", gocui.KeyEsc, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return gocui.ErrQuit
	})
}

func ShowMessageWithOk(g *gocui.Gui, idPrefix, title, message string, width, height int) error {
	maxX, maxY := g.Size()
	x := (maxX - width) / 2
	y := (maxY - height) / 2

	msgId := idPrefix + "_msg"
	okId := idPrefix + "_ok"

	if v, err := g.SetView(msgId, x, y, x+width, y+height, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Title = " " + title + " "
		fmt.Fprintf(v, "\n  %s\n\n", message)
	} else {
		v.Clear()
		v.Title = " " + title + " "
		fmt.Fprintf(v, "\n  %s\n\n", message)
	}

	btnX := x + width - 14
	btnY := y + height - 2
	createButton(g, okId, " OK ", btnX, btnY-1, 10, 2, okId)

	if _, err := g.SetCurrentView(okId); err != nil {
		return err
	}

	g.SetKeybinding(okId, gocui.KeyEnter, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		g.DeleteView(msgId)
		g.DeleteView(okId)
		g.DeleteKeybinding(okId, gocui.KeyEnter, gocui.ModNone)
		g.DeleteKeybinding(msgId, gocui.KeyEsc, gocui.ModNone)
		g.DeleteKeybinding(okId, gocui.KeyEsc, gocui.ModNone)
		if _, err := g.View("inventory"); err == nil {
			g.SetCurrentView("inventory")
		}
		return nil
	})
	g.SetKeybinding(msgId, gocui.KeyEsc, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		g.DeleteView(msgId)
		g.DeleteView(okId)
		g.DeleteKeybinding(okId, gocui.KeyEnter, gocui.ModNone)
		g.DeleteKeybinding(msgId, gocui.KeyEsc, gocui.ModNone)
		g.DeleteKeybinding(okId, gocui.KeyEsc, gocui.ModNone)
		if _, err := g.View("inventory"); err == nil {
			g.SetCurrentView("inventory")
		}
		return nil
	})
	g.SetKeybinding(okId, gocui.KeyEsc, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		g.DeleteView(msgId)
		g.DeleteView(okId)
		g.DeleteKeybinding(okId, gocui.KeyEnter, gocui.ModNone)
		g.DeleteKeybinding(msgId, gocui.KeyEsc, gocui.ModNone)
		g.DeleteKeybinding(okId, gocui.KeyEsc, gocui.ModNone)
		if _, err := g.View("inventory"); err == nil {
			g.SetCurrentView("inventory")
		}
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

func CreateFormField(g *gocui.Gui, labelName, fieldName, labelText string, x, y, w, h int, editable bool, username, seed string, races []string, raceIndex int) error {
	if v, err := g.SetView(labelName, x, y, x+10, y+2, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Frame = false
		fmt.Fprint(v, labelText)
	}

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
		switch fieldName {
		case "username":
			fmt.Fprint(v, username)
		case "seed":
			fmt.Fprint(v, seed)
		case "race":
			fmt.Fprintf(v, "< %s >", races[raceIndex])
		}
	} else if fieldName == "race" {
		v.Clear()
		fmt.Fprintf(v, "< %s >", races[raceIndex])
	}
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

func ShowSimpleDialog(g *gocui.Gui, idPrefix, title, message, buttonText string, width, height int, onClose func()) error {
	maxX, maxY := g.Size()
	x := (maxX - width) / 2
	y := (maxY - height) / 2

	msgId := idPrefix + "_dialog"
	btnId := idPrefix + "_dialog_btn"

	if v, err := g.SetView(msgId, x, y, x+width, y+height, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Frame = true
		v.Title = title
		fmt.Fprintf(v, "\n  %s\n", message)
		fmt.Fprintln(v, "  (Not implemented)")
		fmt.Fprintln(v, "")

		g.SetKeybinding(msgId, gocui.KeyEsc, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
			if onClose != nil {
				onClose()
			}
			g.DeleteView(msgId)
			g.DeleteView(btnId)
			return nil
		})
	}

	buttonY := y + height - 2
	buttonX := x + width - 14
	createButton(g, btnId, buttonText, buttonX, buttonY-1, 12, 2, btnId)

	g.SetKeybinding(btnId, gocui.KeyEnter, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if onClose != nil {
			onClose()
		}
		g.DeleteView(msgId)
		g.DeleteView(btnId)
		return nil
	})

	return nil
}

type ButtonHandler struct {
	ViewName string
	Handler  func(g *gocui.Gui, v *gocui.View) error
}

func HandleMouseClickButtons(g *gocui.Gui, mx, my int, buttons []ButtonHandler) error {
	for _, btn := range buttons {
		if isMouseOver(g, btn.ViewName, mx, my) {
			return btn.Handler(g, nil)
		}
	}
	return nil
}

type ButtonHandlerWithHover struct {
	ViewName   string
	Handler    func(g *gocui.Gui, v *gocui.View) error
	ClearHover func()
}

func HandleMouseClickButtonsWithHover(g *gocui.Gui, mx, my int, buttons []ButtonHandlerWithHover) error {
	for _, btn := range buttons {
		if isMouseOver(g, btn.ViewName, mx, my) {
			if btn.ClearHover != nil {
				btn.ClearHover()
			}
			return btn.Handler(g, nil)
		}
	}
	return nil
}

func EnableMouseAndSetHandler(g *gocui.Gui, handler func(g *gocui.Gui, v *gocui.View) error) {
	g.Mouse = true
	g.SetKeybinding("", gocui.MouseLeft, gocui.ModNone, handler)
}

func ValidateSelectedIndex(selected *int, maxCount int) {
	if *selected < 0 {
		*selected = 0
	}
	if *selected >= maxCount {
		*selected = maxCount - 1
	}
}

func IsValidIndex(selected, maxCount int) bool {
	return selected >= 0 && selected < maxCount
}

func DeleteViews(g *gocui.Gui, viewNames ...string) {
	for _, viewName := range viewNames {
		g.DeleteView(viewName)
	}
}
