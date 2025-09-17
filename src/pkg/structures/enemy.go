package structures

type Enemy struct {
	Entity
	Weapon
	EnemyRace
}

func (enm *Enemy) InflictDamage(Action string, attackedEntity *Entity, spellUsed Spell, multi float64) (int, int) {
	switch Action {
	case "Melee":
		rawDamage := int(float64(enm.EnemyRace.BonusDamage+enm.Weapon.Damage) * multi)
		actualDamage := attackedEntity.TakeDamage(rawDamage)
		return rawDamage, actualDamage
	}
	return 0, 0
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
