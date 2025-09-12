package structures

type Entity struct {
	HP         int
	MaxHP      int
	Damage     int
	Name       string
	Id         int
	Alive      bool
	Level      int
	Helmet     Armors
	Chestplate Armors
	Boots      Armors
}

func (ent *Entity) TakeDamage(damage int) {
	damage -= ent.Helmet.Defense + ent.Chestplate.Defense + ent.Boots.Defense
	if ent.HP-damage <= 0 {
		ent.HP = 0
		ent.Alive = false
	} else {
		ent.HP -= damage
	}
}
