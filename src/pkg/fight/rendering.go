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

	makePlayerBox := func(name string, hp, maxHP, level, mana, defense int, weapon string) []string {
		lines := []string{}
		lines = append(lines, fmt.Sprintf("+%s+", strings.Repeat("-", boxWidth-2)))
		lines = append(lines, fmt.Sprintf("| %-*s|", boxWidth-3, name))
		lines = append(lines, fmt.Sprintf("| %-*s|", boxWidth-3, fmt.Sprintf("Health: %d/%d", hp, maxHP)))
		lines = append(lines, fmt.Sprintf("| %-*s|", boxWidth-3, fmt.Sprintf("Mana: %d", mana)))
		lines = append(lines, fmt.Sprintf("| %-*s|", boxWidth-3, fmt.Sprintf("Level: %d", level)))
		lines = append(lines, fmt.Sprintf("| %-*s|", boxWidth-3, fmt.Sprintf("Defense: %d%%", defense)))
		lines = append(lines, fmt.Sprintf("| %-*s|", boxWidth-3, fmt.Sprintf("Weapon: %s", weapon)))
		lines = append(lines, fmt.Sprintf("+%s+", strings.Repeat("-", boxWidth-2)))
		return lines
	}

	makeEnemyBox := func(name string, hp, maxHP, level, defense int, weapon string) []string {
		lines := []string{}
		lines = append(lines, fmt.Sprintf("+%s+", strings.Repeat("-", boxWidth-2)))
		lines = append(lines, fmt.Sprintf("| %-*s|", boxWidth-3, name))
		lines = append(lines, fmt.Sprintf("| %-*s|", boxWidth-3, fmt.Sprintf("Health: %d/%d", hp, maxHP)))
		lines = append(lines, fmt.Sprintf("| %-*s|", boxWidth-3, "")) // Empty line to match player's mana line

		var levelIndicator string
		if level > 0 {
			levelIndicator = fmt.Sprintf("Depth Lv.%d", level)
		} else {
			levelIndicator = "Surface"
		}
		lines = append(lines, fmt.Sprintf("| %-*s|", boxWidth-3, levelIndicator))
		lines = append(lines, fmt.Sprintf("| %-*s|", boxWidth-3, fmt.Sprintf("Defense: %d%%", defense)))
		lines = append(lines, fmt.Sprintf("| %-*s|", boxWidth-3, fmt.Sprintf("Weapon: %s", weapon)))
		lines = append(lines, fmt.Sprintf("+%s+", strings.Repeat("-", boxWidth-2)))
		return lines
	}

	playerDefenseRaw := player.Helmet.Defense + player.Chestplate.Defense + player.Boots.Defense + structures.GetSetBonusDefense(player.Entity)
	mobDefenseRaw := mob.Helmet.Defense + mob.Chestplate.Defense + mob.Boots.Defense + structures.GetSetBonusDefense(mob.Entity)

	playerDefensePercent := playerDefenseRaw * 2
	if playerDefensePercent > 85 {
		playerDefensePercent = 85
	}
	mobDefensePercent := mobDefenseRaw * 2
	if mobDefensePercent > 85 {
		mobDefensePercent = 85
	}

	playerBox := makePlayerBox(player.Entity.Name, player.HP, player.MaxHP, player.Level, player.Mana, playerDefensePercent, player.Weapon.Name)
	mobBox := makeEnemyBox(mob.Entity.Name, mob.HP, mob.MaxHP, mob.Level, mobDefensePercent, mob.Weapon.Name)

	padding := strings.Repeat(" ", leftPadding)

	fmt.Println()
	fmt.Println(centerText(fmt.Sprintf("===== ROUND %d =====", round)))
	fmt.Println(centerText("YOU ARE IN A FIGHT"))
	fmt.Println()

	for i := range playerBox {
		fmt.Println(padding + playerBox[i] + strings.Repeat(" ", spaceBetween) + mobBox[i])
	}
}
