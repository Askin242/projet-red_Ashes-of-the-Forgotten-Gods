package fight

import (
	"fmt"
	structures "main/pkg/structures"
	"strings"
)

func RenderFight(player *structures.Player, mob *structures.Enemy, playerPlaying bool, round int) {
	boxWidth := 30
	spaceBetween := 8
	screenWidth := 145
	totalWidth := boxWidth*2 + spaceBetween
	if totalWidth > screenWidth {
		screenWidth = totalWidth
	}
	leftPadding := (screenWidth - totalWidth) / 2

	centerText := func(text string) string {
		if len(text) >= screenWidth {
			return text
		}
		padding := (screenWidth - len(text)) / 2
		return strings.Repeat(" ", padding) + text
	}

	makeBox := func(name string, hp, maxHP, level int, weapon string) []string {
		lines := []string{}
		lines = append(lines, fmt.Sprintf("+%s+", strings.Repeat("-", boxWidth-2)))
		lines = append(lines, fmt.Sprintf("| %-*s|", boxWidth-3, name))
		lines = append(lines, fmt.Sprintf("| %-*s|", boxWidth-3, fmt.Sprintf("Health: %d/%d", hp, maxHP)))
		lines = append(lines, fmt.Sprintf("| %-*s|", boxWidth-3, fmt.Sprintf("Level: %d", level)))
		lines = append(lines, fmt.Sprintf("| %-*s|", boxWidth-3, fmt.Sprintf("Weapon: %s", weapon)))
		lines = append(lines, fmt.Sprintf("+%s+", strings.Repeat("-", boxWidth-2)))
		return lines
	}

	playerBox := makeBox(player.Entity.Name, player.HP, player.MaxHP, player.Level, player.Weapon.Name)
	mobBox := makeBox(mob.Entity.Name, mob.HP, mob.MaxHP, mob.Level, mob.Weapon.Name)

	padding := strings.Repeat(" ", leftPadding)

	fmt.Println()
	fmt.Println(centerText(fmt.Sprintf("===== ROUND %d =====", round)))
	fmt.Println(centerText("YOU ARE IN A FIGHT"))
	fmt.Println()

	for i := range playerBox {
		fmt.Println(padding + playerBox[i] + strings.Repeat(" ", spaceBetween) + mobBox[i])
	}
}
