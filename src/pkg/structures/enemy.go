package structures

type Enemy struct {
	Entity
	Weapon
	EnemyRace
}

func (enm *Enemy) InflictDamage(Action string, attackedEntity Entity, spellUsed Spells) {
	switch Action {
	case "Melee":
		damageOutput := enm.EnemyRace.BonusDamage + enm.Weapon.Damage
		attackedEntity.TakeDamage(damageOutput)
	}
}

func InitEnemy(name string, race string) Enemy {
	return Enemy{
		Entity: Entity{
			HP:         100,
			MaxHP:      100,
			Name:       name,
			Alive:      true,
			Level:      0,
			Helmet:     GetRandomArmorByType("Helmet"),
			Chestplate: GetRandomArmorByType("Chestplate"),
			Boots:      GetRandomArmorByType("Boots"),
		},
		Weapon:    AllWeapons["Sword"],
		EnemyRace: AllEnemyRaces[race],
	}
}
