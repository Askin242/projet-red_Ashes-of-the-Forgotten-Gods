package fight

import (
	"fmt"
	ui "main/pkg/ui"
	"os"
	"time"

	"golang.org/x/term"
)

func QuickTimeEvent(speed time.Duration, length int) float64 {
	ui.EnableAnsiColors()
	if length < 5 {
		length = 5
	}

	xPos := length / 2
	oPos := 0
	direction := 1

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	ticker := time.NewTicker(speed)
	defer ticker.Stop()

	render := func() {
		fmt.Print("\r")
		for i := 0; i < length; i++ {
			if i == oPos {
				fmt.Print("\033[33mo\033[0m") // Strange char are ansi codes for colors
			} else if i == xPos {
				fmt.Print("\033[32;1mx\033[0m")
			} else if abs(i-xPos) <= 2 {
				fmt.Print("\033[32m-\033[0m")
			} else {
				fmt.Print("-")
			}
		}
		fmt.Print("  (Press SPACE when 'o' is on 'x' for PERFECT block! Green zone = good block)")
	}

	render()

	input := make(chan byte)
	go func() {
		buf := make([]byte, 1)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil || n == 0 {
				close(input)
				return
			}
			input <- buf[0]
		}
	}()

	for {
		select {
		case <-ticker.C:
			oPos += direction
			if oPos == length-1 || oPos == 0 {
				direction *= -1
			}
			render()

		case b, ok := <-input:
			if !ok {
				fmt.Println()
				return 1.0
			}
			if b == ' ' {
				fmt.Println()
				if oPos == xPos {
					fmt.Println("\033[32;1mYou perfectly blocked the attack!\033[0m")
					return 0
				} else if abs(oPos-xPos) <= 2 {
					fmt.Println("\033[32mYou did a good block.\033[0m")
					return 0.4
				} else {
					fmt.Println("\033[31mYou missed the block.\033[0m")
					return 1.0
				}
			}
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
