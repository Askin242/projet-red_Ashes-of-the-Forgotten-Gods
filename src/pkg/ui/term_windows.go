//go:build windows

package ui

import (
	"fmt"
	"os"
	"os/exec"

	"golang.org/x/sys/windows"
)

func enableAnsiColors() {
	var originalMode uint32
	stdout := windows.Handle(os.Stdout.Fd())
	windows.GetConsoleMode(stdout, &originalMode)
	windows.SetConsoleMode(stdout, originalMode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
}

func setTerminalSize(cols, rows int) {
	cmd := exec.Command("cmd", "/C", fmt.Sprintf("mode con: cols=%d lines=%d", cols, rows))
	cmd.Run()
}
