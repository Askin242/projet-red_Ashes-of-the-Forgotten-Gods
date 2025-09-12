package structures

type Player struct {
	Entity
	Weapons
	Races
	Helmet     Armors
	Chestplate Armors
	Boots      Armors
	Mana       int
	Money      int
}

func (plr *Player) TakeDamage(damage int) {
	if plr.Entity.HP-damage <= 0 {
		plr.Entity.HP = 0
		plr.Alive = false
	} else {
		plr.Entity.HP -= damage
	}
}

func (plr *Player) InflictDamage(Action string, attackedEntity Entity, spellUsed Spells) {
	switch Action {
	case "Melee":
		damageOutput := plr.Races.BonusDamage + plr.Weapons.Damage
		attackedEntity.TakeDamage(damageOutput)

	case "Spell":
		if spellUsed.Cost >= plr.Mana {
			plr.Mana -= spellUsed.Cost
			damageOutput := plr.Races.BonusDamage + spellUsed.Damage
			attackedEntity.TakeDamage(damageOutput)
		}
	}
}

func InitCharacter(username, race string) Player {
	mainPlayer := Player{
		Entity: Entity{
			HP:     100,
			MaxHP:  100,
			Damage: 10,
			Name:   username,
			Id:     123,
			Alive:  true,
			Level:  0,
		},
		Weapons: Weapons{
			Damage: 5,
			Name:   "Yapper",
			Id:     12345,
		},
		Races: Races{
			Name:        race,
			BonusHP:     AllRaces[race].BonusHP,
			BonusDamage: AllRaces[race].BonusDamage,
			BonusMana:   AllRaces[race].BonusMana,
			Skill:       AllSpells[AllRaces[race].Skill.Name],
		},
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
		Mana:  100,
		Money: 100,
	}
	mainPlayer.Mana += mainPlayer.Races.BonusMana
	mainPlayer.MaxHP += mainPlayer.Races.BonusHP
	mainPlayer.HP = mainPlayer.MaxHP

	return mainPlayer
}
