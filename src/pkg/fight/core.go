package fight

import (
	"bufio"
	"fmt"
	structures "main/pkg/structures"

	ui "main/pkg/ui"
	"math/rand"
	"os"
	"strings"
	"time"
)

func flushInput(reader *bufio.Reader) {
	for {
		if reader.Buffered() == 0 {
			break
		}
		_, _ = reader.ReadString('\n')
	}
}

func readLine(reader *bufio.Reader) string {
	for reader.Buffered() > 0 {
		_, _ = reader.ReadByte()
	}

	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func getStartingPlayer(player *structures.Player, enemy *structures.Enemy) bool {
	if player.Entity.Initiative == enemy.Entity.Initiative {
		return true
	}
	return player.Entity.Initiative > enemy.Entity.Initiative
}

func StartFight(character *structures.Player, enemy *structures.Enemy) {
	fmt.Printf("%v has come across the malicious %v!\n", character.Entity.Name, enemy.Entity.Name)
	fmt.Println("Determining who will start...")

	playerTurn := getStartingPlayer(character, enemy)
	if playerTurn {
		fmt.Printf("%v will start the fight!\n\n", character.Entity.Name)
	} else {
		fmt.Printf("%v will start the fight!\n\n", enemy.Entity.Name)
	}

	reader := bufio.NewReader(os.Stdin)
	roundNumber := 0
	speed := 150 * time.Millisecond

	time.Sleep(4 * time.Second)
	ui.CallClear()

	for character.Alive && enemy.Entity.Alive {
		roundNumber++

		RenderFight(character, enemy, playerTurn, roundNumber)
		structures.ProcessEffects(&character.Entity)
		structures.ProcessEffects(&enemy.Entity)

		if playerTurn {
			chosen := false
			for !chosen {
				fmt.Println("[1] Attack with your weapon")
				fmt.Println("[2] Use your Spell")
				flushInput(reader)
				fmt.Print("> ")

				mode := readLine(reader)

				switch mode {
				case "1":
					damage := character.InflictDamage("Melee", &enemy.Entity, structures.AllSpells["None"], 1.0)
					fmt.Printf("[%s] used their weapon dealing %d damage to [%s]!\n",
						character.Entity.Name, damage, enemy.Entity.Name)
					chosen = true
					time.Sleep(2 * time.Second)
					ui.CallClear()

				case "2":
					for i, spell := range character.Spells {
						fmt.Printf("[%d] %s (Cost: %d Mana, Damage: %d, Element: %s)\n",
							i+1, spell.Name, spell.Cost, spell.Damage, spell.Element)
					}
					flushInput(reader)
					fmt.Print("> ")

					spellChoice := readLine(reader)

					if spellChoice == "" {
						RenderFight(character, enemy, playerTurn, roundNumber)
						continue
					}

					spellIndex := 0
					fmt.Sscanf(spellChoice, "%d", &spellIndex)

					if spellIndex > 0 && spellIndex <= len(character.Spells) {
						chosenSpell := character.Spells[spellIndex-1]
						damage := character.InflictDamage("Spell", &enemy.Entity, chosenSpell, 1.0)
						fmt.Println(damage)
						fmt.Printf("[%s] used %s dealing %d damage to [%s]!\n",
							character.Entity.Name, chosenSpell.Name, damage, enemy.Entity.Name)
						chosen = true
						time.Sleep(2 * time.Second)
						ui.CallClear()
					} else {
						fmt.Println("Invalid spell choice.")
						RenderFight(character, enemy, playerTurn, roundNumber)
					}

				default:
					fmt.Println("Invalid input! Please choose again.")
					RenderFight(character, enemy, playerTurn, roundNumber)
				}
			}
		} else {
			fmt.Println("\n!!! Incoming attack !!!")
			damageMultiplier := QuickTimeEvent(speed, 12)
			damage := enemy.InflictDamage("Melee", &character.Entity, structures.AllSpells["None"], damageMultiplier)
			fmt.Printf("[%s] attacked [%s] dealing %d damage!\n",
				enemy.Entity.Name, character.Entity.Name, damage)
		}

		playerTurn = !playerTurn
	}

	if character.Alive {
		fmt.Printf("\n%s has defeated %s!\n", character.Entity.Name, enemy.Entity.Name)
		loot := structures.GenerateLootFromEnemy(enemy.EnemyRace)
		character.AddItem(loot)
		droppedMoney := rand.Intn(5) + 1
		character.Money += droppedMoney
		fmt.Printf("%s found a %s, aswell as %d coins!\n", character.Entity.Name, loot.GetItem().Name, droppedMoney)
		character.AddXP(character.GetxpFromMob(enemy.Entity))
	} else {
		fmt.Printf("\n%s has been defeated by %s!\n", character.Entity.Name, enemy.Entity.Name)
	}
}
