package structures

type Weapon struct {
	Damage int
	Name   string
	Id     int
}

var (
	Sword = Weapon{
		Name:   "Sword",
		Damage: 10,
		Id:     1,
	}
	Axe = Weapon{
		Name:   "Axe",
		Damage: 13,
		Id:     2,
	}
	DoubleAxes = Weapon{
		Name:   "DoubleAxes",
		Damage: 18,
		Id:     3,
	}
	Spear = Weapon{
		Name:   "Spear",
		Damage: 25,
		Id:     4,
	}
)

var AllWeapons = map[string]Weapon{
	"Sword":      Sword,
	"Axe":        Axe,
	"DoubleAxes": DoubleAxes,
	"Spear":      Spear,
}
