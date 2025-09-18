package structures

import "math/rand"

type Enemy struct {
	Entity
	Weapon
	EnemyRace
	IsBoss bool
	Mana   int
	Spells []Spell
}

func (enm *Enemy) InflictDamage(Action string, attackedEntity *Entity, spellUsed Spell, multi float64) (int, int) {
	switch Action {
	case "Melee":
		rawDamage := int(float64(enm.EnemyRace.BonusDamage+enm.Weapon.Damage) * multi)
		actualDamage := attackedEntity.TakeDamage(rawDamage)
		return rawDamage, actualDamage
	case "Spell":
		if spellUsed.Cost <= enm.Mana {
			enm.Mana -= spellUsed.Cost
			rawDamage := int(float64(enm.EnemyRace.BonusDamage+spellUsed.Damage) * multi)
			actualDamage := attackedEntity.TakeDamage(rawDamage)
			ApplySpellEffect(spellUsed, attackedEntity)
			return rawDamage, actualDamage
		}
	case "HeavySlam":
		base := enm.EnemyRace.BonusDamage + enm.Weapon.Damage
		rawDamage := int(float64(base) * 1.8 * multi)
		actualDamage := attackedEntity.TakeDamage(rawDamage)
		return rawDamage, actualDamage
	}
	return 0, 0
}

func InitEnemy(name string, race string) Enemy {
	enemy := Enemy{
		Entity: Entity{
			HP:         100,
			MaxHP:      100,
			Name:       name,
			Alive:      true,
			Level:      1,
			Helmet:     GetRandomArmorByType("Helmet"),
			Chestplate: GetRandomArmorByType("Chestplate"),
			Boots:      GetRandomArmorByType("Boots"),
			Initiative: 10,
			defaultXP:  rand.Intn(10) + 1,
		},
		Weapon:    AllWeapons["Sword"],
		EnemyRace: AllEnemyRaces[race],
		IsBoss:    false,
		Mana:      0,
		Spells:    []Spell{},
	}

	enemy.Entity.Initiative += enemy.EnemyRace.BonusInitiative
	return enemy
}

func InitBoss(name string, race string) Enemy {
	boss := Enemy{
		Entity: Entity{
			HP:         220,
			MaxHP:      220,
			Name:       name,
			Alive:      true,
			Level:      5,
			Helmet:     AllHelmets["SunBreaker"],
			Chestplate: AllChestplates["SunBreaker"],
			Boots:      AllBoots["SunBreaker"],
			Initiative: 99,
			defaultXP:  rand.Intn(40) + 100,
		},
		Weapon:    AllWeapons["Axe"],
		EnemyRace: AllEnemyRaces[race],
		IsBoss:    true,
		Mana:      150,
	}
	pool := []Spell{}
	for _, s := range AllSpells {
		if s.Name != "None" && s.Name != "HandPunch" {
			pool = append(pool, s)
		}
	}
	r := GetRNG()
	for i := range pool {
		j := r.Intn(i + 1)
		pool[i], pool[j] = pool[j], pool[i]
	}
	if len(pool) > 3 {
		boss.Spells = pool[:3]
	} else {
		boss.Spells = pool
	}
	return boss
}

// InitScaledEnemy creates an enemy with stats scaled based on dungeon level
func InitScaledEnemy(name string, race string, dungeonLevel int) Enemy {
	enemy := InitEnemy(name, race)

	// Calculate scaling factor based on dungeon level
	// Level 0 = 1.0x (base difficulty)
	// Level 1 = 1.2x
	// Level 2 = 1.4x
	// level 3 is a boss level
	scalingFactor := 1.0 + (float64(dungeonLevel) * 0.2)

	baseHP := 100 + enemy.EnemyRace.BonusHP
	scaledHP := int(float64(baseHP) * scalingFactor)
	enemy.Entity.HP = scaledHP
	enemy.Entity.MaxHP = scaledHP

	enemy.EnemyRace.BonusDamage = int(float64(enemy.EnemyRace.BonusDamage) * scalingFactor)

	// Set level for display/identification
	enemy.Entity.Level = dungeonLevel

	if dungeonLevel > 0 {
		// Higher level enemies get better armor
		armorScaling := 1.0 + (float64(dungeonLevel) * 0.2)
		enemy.Entity.Helmet.Defense = int(float64(enemy.Entity.Helmet.Defense) * armorScaling)
		enemy.Entity.Chestplate.Defense = int(float64(enemy.Entity.Chestplate.Defense) * armorScaling)
		enemy.Entity.Boots.Defense = int(float64(enemy.Entity.Boots.Defense) * armorScaling)
	}

	// Give higher level enemies better weapons occasionally
	if dungeonLevel >= 2 {
		weapons := []string{"Axe", "DoubleAxes", "Spear"}
		weaponIndex := GetRNG().Intn(len(weapons))
		selectedWeapon := AllWeapons[weapons[weaponIndex]]

		// Scale weapon damage
		weaponScaling := 1.0 + (float64(dungeonLevel) * 0.15)
		selectedWeapon.Damage = int(float64(selectedWeapon.Damage) * weaponScaling)
		enemy.Weapon = selectedWeapon
	}

	return enemy
}
