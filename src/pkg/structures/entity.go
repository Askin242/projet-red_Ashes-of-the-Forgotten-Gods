package structures

import "fmt"

type Effect struct {
	Name     string
	Duration int
	Modifier float64
}

type Entity struct {
	HP         int
	MaxHP      int
	Name       string
	Alive      bool
	Level      int
	Initiative int
	Helmet     Armors
	Chestplate Armors
	Boots      Armors
	defaultXP  int
	Effects    []Effect
}

func (ent *Entity) TakeDamage(damage int) int {
	defense := ent.Helmet.Defense + ent.Chestplate.Defense + ent.Boots.Defense + GetSetBonusDefense(*ent)

	defensePercent := float64(defense) * 2.0
	if defensePercent > 85 {
		defensePercent = 85
	}

	actualDamage := int(float64(damage) * (100.0 - defensePercent) / 100.0)

	if damage > 0 && actualDamage == 0 {
		actualDamage = 1
	}

	if ent.HP-actualDamage <= 0 {
		ent.HP = 0
		ent.Alive = false
	} else {
		ent.HP -= actualDamage
	}
	return actualDamage
}

func ProcessEffects(entity *Entity) {
	remainingEffects := []Effect{}
	for _, eff := range entity.Effects {
		switch eff.Name {
		case "Burn":
			burnDmg := int(float64(entity.MaxHP) * eff.Modifier)
			entity.TakeDamage(burnDmg)
			fmt.Printf("%s takes %d burn damage!\n", entity.Name, burnDmg)
		}
		eff.Duration--
		if eff.Duration > 0 {
			remainingEffects = append(remainingEffects, eff)
		}
	}
	entity.Effects = remainingEffects
}
