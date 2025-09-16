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
	Effects    []Effect
}

func (ent *Entity) TakeDamage(damage int) {
	defense := ent.Helmet.Defense + ent.Chestplate.Defense + ent.Boots.Defense + GetSetBonusDefense(*ent)
	damage -= defense
	if ent.HP-damage <= 0 {
		ent.HP = 0
		ent.Alive = false
	} else {
		ent.HP -= damage
	}
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
