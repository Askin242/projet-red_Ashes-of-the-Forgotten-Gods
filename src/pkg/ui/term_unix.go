//go:build !windows

package ui

import "fmt"

func enableAnsiColors() {
	// Modern terminals support ANSI by default on Unix-like systems
}

func setTerminalSize(cols, rows int) {
	fmt.Printf("\033[8;%d;%dt", rows, cols)
}
