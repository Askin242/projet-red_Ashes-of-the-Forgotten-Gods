package structures

type Player struct {
	Entity
	Weapon
	Race
	Mana           int
	Money          int
	Inventory      []InventoryEntry
	MaxCarryWeight int
}

func (plr *Player) InflictDamage(Action string, attackedEntity Entity, spellUsed Spells) {
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

func InitCharacter(username, race string) Player {
	mainPlayer := Player{
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
		},
		Weapon:         AllWeapons["Sword"],
		Race:           AllRaces[race],
		Mana:           100,
		Money:          100,
		Inventory:      []InventoryEntry{},
		MaxCarryWeight: 10,
	}
	mainPlayer.Mana += mainPlayer.Race.BonusMana
	mainPlayer.MaxHP += mainPlayer.Race.BonusHP
	mainPlayer.HP = mainPlayer.MaxHP

	mainPlayer.AddItem(AllPotions["Poison"])
	mainPlayer.AddItem(AllPotions["Poison"])
	mainPlayer.AddItem(AllPotions["Poison"])

	return mainPlayer
}
