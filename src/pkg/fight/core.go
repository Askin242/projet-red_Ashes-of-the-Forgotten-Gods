package fight

import (
	"bufio"
	"fmt"
	structures "main/pkg/structures"
	ui "main/pkg/ui"
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
	speed := 20 * time.Millisecond

	time.Sleep(4 * time.Second)
	ui.ClearScreen()

	for character.Alive && enemy.Entity.Alive {
		roundNumber++

		if playerTurn {
			maxMana := 100 + character.Race.BonusMana
			if character.Mana < maxMana {
				character.Mana += 10
				if character.Mana > maxMana {
					character.Mana = maxMana
				}
			}
		}

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
					rawDamage, actualDamage := character.InflictDamage("Melee", &enemy.Entity, structures.AllSpells["None"], 1.0)
					fmt.Printf("[%s] used their weapon dealing %d damage (%d before defense) to [%s]!\n",
						character.Entity.Name, actualDamage, rawDamage, enemy.Entity.Name)
					chosen = true
					time.Sleep(2 * time.Second)
					ui.ClearScreen()

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
						rawDamage, actualDamage := character.InflictDamage("Spell", &enemy.Entity, chosenSpell, 1.0)
						fmt.Printf("[%s] used %s dealing %d damage (%d before defense) to [%s]!\n",
							character.Entity.Name, chosenSpell.Name, actualDamage, rawDamage, enemy.Entity.Name)
						chosen = true
						time.Sleep(2 * time.Second)
						ui.ClearScreen()
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
			fmt.Println("Quick Time Event: Time your block to reduce incoming damage!")
			damageMultiplier := QuickTimeEvent(speed, 20)

			baseDamage := enemy.EnemyRace.BonusDamage + enemy.Weapon.Damage
			blockedDamage := int(float64(baseDamage) * (1.0 - damageMultiplier))

			rawDamage, actualDamage := enemy.InflictDamage("Melee", &character.Entity, structures.AllSpells["None"], damageMultiplier)

			if blockedDamage > 0 {
				fmt.Printf("[%s] attacked [%s] dealing %d damage (%d base damage, %d blocked by timing, %d reduced by armor)!\n",
					enemy.Entity.Name, character.Entity.Name, actualDamage, baseDamage, blockedDamage, rawDamage-actualDamage)
			} else {
				fmt.Printf("[%s] attacked [%s] dealing %d damage (%d base damage, %d reduced by armor)!\n",
					enemy.Entity.Name, character.Entity.Name, actualDamage, baseDamage, rawDamage-actualDamage)
			}
		}

		playerTurn = !playerTurn
	}

	if character.Alive {
		fmt.Printf("\n%s has defeated %s!\n", character.Entity.Name, enemy.Entity.Name)
		loot := structures.GenerateLootFromEnemy(enemy.EnemyRace)
		character.AddItem(loot)
		droppedMoney := structures.GetRNG().Intn(5) + 1
		character.Money += droppedMoney
		fmt.Printf("%s found a %s, aswell as %d coins!\n", character.Entity.Name, loot.GetItem().Name, droppedMoney)

		structures.RefreshSeedState()
	} else {
		fmt.Printf("\n%s has been defeated by %s!\n", character.Entity.Name, enemy.Entity.Name)
	}
}
