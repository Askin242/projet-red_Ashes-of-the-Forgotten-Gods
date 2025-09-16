package structures

type Enemy struct {
	Entity
	Weapon
	EnemyRace
}

func (enm *Enemy) InflictDamage(Action string, attackedEntity *Entity, spellUsed Spell, multi float64) int {
	switch Action {
	case "Melee":
		damageOutput := int(float64(enm.EnemyRace.BonusDamage+enm.Weapon.Damage) * multi)
		attackedEntity.TakeDamage(damageOutput)
		return damageOutput
	}
	return 0
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
			Initiative: 0,
		},
		Weapon:    AllWeapons["Sword"],
		EnemyRace: AllEnemyRaces[race],
	}
}
