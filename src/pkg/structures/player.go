package structures

import (
	"fmt"
	"main/pkg/save"
)

type Player struct {
	Entity
	Weapon
	Race
	Mana           int
	Money          int
	Inventory      Inventory
	MaxCarryWeight int
	Spells         []Spell
}

func ApplySpellEffect(spell Spell, target *Entity) {
	switch spell.Element {
	case "Fire":
		target.Effects = append(target.Effects, Effect{
			Name:     "Burn",
			Duration: 3,
			Modifier: 0.04, // 4% HP per turn
		})

	case "Poison":
		target.Effects = append(target.Effects, Effect{
			Name:     "Poisoned",
			Duration: 2,
			Modifier: 0.4, // Reduce defense by 40%
		})

	case "Lightning":
		target.Effects = append(target.Effects, Effect{
			Name:     "Shocked",
			Duration: 3,
			Modifier: 0.1, // 10% chance to miss
		})

	case "Ice":
		target.Effects = append(target.Effects, Effect{
			Name:     "Frozen",
			Duration: 3,
			Modifier: 0.75, // 25% damage reduction
		})
	}
}

func (plr *Player) InflictDamage(action string, attackedEntity *Entity, spellUsed Spell, damageMultiplier float64) int {
	switch action {
	case "Melee":
		damageOutput := int(float64(plr.Race.BonusDamage+plr.Weapon.Damage) * damageMultiplier)
		attackedEntity.TakeDamage(damageOutput)
		return damageOutput
	case "Spell":
		if spellUsed.Cost <= plr.Mana {
			plr.Mana -= spellUsed.Cost
			damageOutput := int(float64(plr.Race.BonusDamage+spellUsed.Damage) * damageMultiplier)
			attackedEntity.TakeDamage(damageOutput)
			ApplySpellEffect(spellUsed, attackedEntity)
			return damageOutput
		} else {
			fmt.Println("Not enough mana!")
		}
	}
	return 0
}

func (plr *Player) LevelUp() int {
	plr.Level++
	plr.MaxHP += 10
	plr.HP += 10
	plr.Mana += 10
	plr.Race.BonusDamage += 10
	return plr.Level
}

func (plr *Player) CurrentCarryWeight() int {
	total := 0
	for _, entry := range plr.Inventory {
		total += entry.GetItem().Weight
	}
	return total
}

func (plr *Player) CanAddItem(entry InventoryEntry) bool {
	return plr.CurrentCarryWeight()+entry.GetItem().Weight < plr.MaxCarryWeight // Not <= to allow to add materials (0weight) even if full
}

func (plr *Player) AddItem(entry InventoryEntry) bool {
	if plr.CanAddItem(entry) {
		plr.Inventory = append(plr.Inventory, entry)
		return true
	}
	return false
}

func (plr *Player) RemoveItem(entry InventoryEntry) bool {
	for i, item := range plr.Inventory {
		if item.GetItem().Id == entry.GetItem().Id {
			plr.Inventory = append(plr.Inventory[:i], plr.Inventory[i+1:]...)
			return true
		}
	}
	return false
}

func (plr *Player) CountMaterial(materialName string) int {
	count := 0
	for _, entry := range plr.Inventory {
		if m, ok := entry.(Material); ok {
			if m.Key == materialName {
				count++
			}
		}
	}
	return count
}

func (plr *Player) RemoveMaterials(materialName string, amount int) int {
	removed := 0
	for i := 0; i < len(plr.Inventory) && removed < amount; i++ {
		if m, ok := plr.Inventory[i].(Material); ok && m.Key == materialName {
			plr.Inventory = append(plr.Inventory[:i], plr.Inventory[i+1:]...)
			removed++
			i--
		}
	}
	return removed
}

// Batch material utilities for crafting
func (plr *Player) HasMaterialsBatch(req map[string]int) bool {
	for key, amt := range req {
		if plr.CountMaterial(key) < amt {
			return false
		}
	}
	return true
}

func (plr *Player) RemoveMaterialsBatch(req map[string]int) bool {
	if !plr.HasMaterialsBatch(req) {
		return false
	}
	for key, amt := range req {
		removed := plr.RemoveMaterials(key, amt)
		if removed < amt {
			return false
		}
	}
	return true
}

func (plr *Player) UpgradeInventorySlot() {
	plr.MaxCarryWeight += 10
}

func (plr *Player) UsePotion(p Potion) bool {
	for _, entry := range plr.Inventory {
		item := entry.GetItem()
		if item.Id == p.Item.Id {
			switch p.Type {
			case "Heal":
				healAmount := 50 * p.Size
				plr.HP += healAmount
				if plr.HP > plr.MaxHP {
					plr.HP = plr.MaxHP
				}
			default:
			}
			plr.RemoveItem(entry)
			return true
		}
	}
	return false
}

func InitCharacter(username, race string) Player {
	mainPlayer := Player{}
	err := save.LoadAny("player", &mainPlayer)
	if err != nil {
		mainPlayer = Player{
			Entity: Entity{
				HP:    100,
				MaxHP: 100,
				Name:  username,
				Alive: true,
				Level: 0,
				Helmet: Armors{
					Name:    "None",
					Type:    "Helmet",
					Defense: 0,
				},
				Chestplate: Armors{
					Name:    "None",
					Type:    "Chestplate",
					Defense: 0,
				},
				Boots: Armors{
					Name:    "None",
					Type:    "Boots",
					Defense: 0,
				},
				Initiative: 0,
			},
			Weapon:         AllWeapons["Sword"],
			Race:           AllRaces[race],
			Mana:           100,
			Money:          100,
			Inventory:      Inventory{},
			MaxCarryWeight: 10,
			Spells:         []Spell{AllSpells["HandPunch"]},
		}
		mainPlayer.Mana += mainPlayer.Race.BonusMana
		mainPlayer.MaxHP += mainPlayer.Race.BonusHP
		mainPlayer.HP = mainPlayer.MaxHP

		for range "123" {
			mainPlayer.AddItem(GetPotion("Heal", 1, 0))
		}
		save.SaveAny("player", mainPlayer)
	}

	return mainPlayer
}
