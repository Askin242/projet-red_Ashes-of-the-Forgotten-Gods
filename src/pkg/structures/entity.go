package structures

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
