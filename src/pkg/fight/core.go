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
	playerInitiative := player.Entity.Initiative
	enemyInitiative := enemy.Entity.Initiative

	playerRoll := structures.GetRNG().Intn(10) + 1
	enemyRoll := structures.GetRNG().Intn(10) + 1

	playerTotal := playerInitiative + playerRoll
	enemyTotal := enemyInitiative + enemyRoll

	return playerTotal >= enemyTotal
}

func StartFight(character *structures.Player, enemy *structures.Enemy) bool {
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
	speed := 30 * time.Millisecond

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
			action := "Melee"
			chosenSpell := structures.AllSpells["None"]
			if enemy.IsBoss {
				if enemy.Mana < 200 {
					enemy.Mana += 15
					if enemy.Mana > 200 {
						enemy.Mana = 200
					}
				}
				availableSpells := []structures.Spell{}
				for _, sp := range enemy.Spells {
					if sp.Cost <= enemy.Mana {
						availableSpells = append(availableSpells, sp)
					}
				}
				r := structures.GetRNG()
				if len(availableSpells) > 0 {
					roll := r.Intn(100)
					if roll < 50 {
						action = "Spell"
						total := 0
						for _, sp := range availableSpells {
							total += sp.Damage
						}
						pick := r.Intn(total)
						sacc := 0
						for _, sp := range availableSpells {
							sacc += sp.Damage
							if pick < sacc {
								chosenSpell = sp
								break
							}
						}
					} else if roll < 80 {
						action = "HeavySlam"
					} else {
						action = "Melee"
					}
				} else {
					if r.Intn(100) < 60 {
						action = "HeavySlam"
					} else {
						action = "Melee"
					}
				}
			}

			fmt.Println("\n!!! Incoming attack !!!")
			fmt.Println("Quick Time Event: Perfect timing blocks 100% damage, good timing blocks 40%!")
			damageMultiplier := QuickTimeEvent(speed, 20)

			baseDamage := enemy.EnemyRace.BonusDamage + enemy.Weapon.Damage
			if action == "Spell" {
				baseDamage = enemy.EnemyRace.BonusDamage + chosenSpell.Damage
			} else if action == "HeavySlam" {
				baseDamage = int(float64(baseDamage) * 1.8)
			}
			blockedDamage := int(float64(baseDamage) * (1.0 - damageMultiplier))

			rawDamage, actualDamage := enemy.InflictDamage(action, &character.Entity, chosenSpell, damageMultiplier)

			switch action {
			case "Spell":
				if damageMultiplier == 0.0 {
					fmt.Printf("[%s] cast %s on [%s] but it was PERFECTLY BLOCKED!\n", enemy.Entity.Name, chosenSpell.Name, character.Entity.Name)
				} else if blockedDamage > 0 {
					fmt.Printf("[%s] cast %s on [%s] dealing %d damage (%d base damage, %d blocked by timing, %d reduced by armor)!\n",
						enemy.Entity.Name, chosenSpell.Name, character.Entity.Name, actualDamage, baseDamage, blockedDamage, rawDamage-actualDamage)
				} else {
					fmt.Printf("[%s] cast %s on [%s] dealing %d damage (%d base damage, %d reduced by armor)!\n",
						enemy.Entity.Name, chosenSpell.Name, character.Entity.Name, actualDamage, baseDamage, rawDamage-actualDamage)
				}
			case "HeavySlam":
				if damageMultiplier == 0.0 {
					fmt.Printf("[%s] used a HEAVY SLAM on [%s] but it was PERFECTLY BLOCKED!\n", enemy.Entity.Name, character.Entity.Name)
				} else if blockedDamage > 0 {
					fmt.Printf("[%s] used a HEAVY SLAM on [%s] dealing %d damage (%d base damage, %d blocked by timing, %d reduced by armor)!\n",
						enemy.Entity.Name, character.Entity.Name, actualDamage, baseDamage, blockedDamage, rawDamage-actualDamage)
				} else {
					fmt.Printf("[%s] used a HEAVY SLAM on [%s] dealing %d damage (%d base damage, %d reduced by armor)!\n",
						enemy.Entity.Name, character.Entity.Name, actualDamage, baseDamage, rawDamage-actualDamage)
				}
			default:
				if damageMultiplier == 0.0 {
					fmt.Printf("[%s] attacked [%s] but the attack was PERFECTLY BLOCKED! No damage taken!\n",
						enemy.Entity.Name, character.Entity.Name)
				} else if blockedDamage > 0 {
					fmt.Printf("[%s] attacked [%s] dealing %d damage (%d base damage, %d blocked by timing, %d reduced by armor)!\n",
						enemy.Entity.Name, character.Entity.Name, actualDamage, baseDamage, blockedDamage, rawDamage-actualDamage)
				} else {
					fmt.Printf("[%s] attacked [%s] dealing %d damage (%d base damage, %d reduced by armor)!\n",
						enemy.Entity.Name, character.Entity.Name, actualDamage, baseDamage, rawDamage-actualDamage)
				}
			}
		}

		playerTurn = !playerTurn
	}

	if character.Alive {
		fmt.Printf("\n%s has defeated %s!\n", character.Entity.Name, enemy.Entity.Name)
		loot := structures.GenerateLootFromEnemy(enemy.EnemyRace)
		character.AddItem(loot)
		droppedMoney := structures.GetRNG().Intn(30) + 1
		character.Money += droppedMoney
		fmt.Printf("%s found a %s, aswell as %d coins!\n", character.Entity.Name, loot.GetItem().Name, droppedMoney)
		structures.RefreshSeedState()
		character.AddXP(character.GetxpFromMob(enemy.Entity))
	} else {
		fmt.Printf("\n%s has been defeated by %s!\n", character.Entity.Name, enemy.Entity.Name)
	}
	return character.Alive
}
