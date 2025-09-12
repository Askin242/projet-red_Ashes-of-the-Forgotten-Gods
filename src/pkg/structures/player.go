package structures

type Player struct {
	Entity
	Weapon
	Races
	Mana  int
	Money int
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
		damageOutput := plr.Races.BonusDamage + plr.Weapon.Damage
		attackedEntity.TakeDamage(damageOutput)

	case "Spell":
		if spellUsed.Cost >= plr.Mana {
			plr.Mana -= spellUsed.Cost
			damageOutput := plr.Races.BonusDamage + spellUsed.Damage
			attackedEntity.TakeDamage(damageOutput)
		}
	}
}

func (plr *Player) LevelUp() int {
	plr.Level++
	plr.MaxHP += 10
	plr.HP += 10
	plr.Mana += 10
	plr.Entity.Damage += 10
	return plr.Level
}

func InitCharacter(username, race string) Player {
	mainPlayer := Player{
		Entity: Entity{
			HP:     100,
			MaxHP:  100,
			Damage: 0,
			Name:   username,
			Id:     0,
			Alive:  true,
			Level:  0,
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
		Weapon: AllWeapons["Sword"],
		Races:  AllRaces[race],
		Mana:   100,
		Money:  100,
	}
	mainPlayer.Mana += mainPlayer.Races.BonusMana
	mainPlayer.MaxHP += mainPlayer.Races.BonusHP
	mainPlayer.HP = mainPlayer.MaxHP

	return mainPlayer
}
