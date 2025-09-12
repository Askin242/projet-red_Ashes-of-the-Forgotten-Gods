package structures

type Entity struct {
	HP     int
	MaxHP  int
	Damage int
	Name   string
	Id     int
	Alive  bool
	Level  int
}

func (ent *Entity) TakeDamage(damage int) {
	if ent.HP-damage <= 0 {
		ent.HP = 0
		ent.Alive = false
	} else {
		ent.HP -= damage
	}
}
