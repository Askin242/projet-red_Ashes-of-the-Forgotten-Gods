package structures

import (
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

func (plr *Player) InflictDamage(Action string, attackedEntity *Entity, spellUsed Spell) {
	switch Action {
	case "Melee":
		damageOutput := plr.Race.BonusDamage + plr.Weapon.Damage
		attackedEntity.TakeDamage(damageOutput)

	case "Spell":
		if spellUsed.Cost >= plr.Mana {
			plr.Mana -= spellUsed.Cost
			damageOutput := plr.Race.BonusDamage + spellUsed.Damage
			attackedEntity.TakeDamage(damageOutput)
		}
	}
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
	return plr.CurrentCarryWeight()+entry.GetItem().Weight <= plr.MaxCarryWeight
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

func InitCharacter(username, race, saveId string) Player {
	mainPlayer := Player{}
	err := save.LoadAny(saveId, "player", &mainPlayer)
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
		save.SaveAny(saveId, "player", mainPlayer)
	}

	return mainPlayer
}
