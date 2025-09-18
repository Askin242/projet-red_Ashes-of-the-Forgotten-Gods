package structures

type Weapon struct {
	Damage int
	Name   string
	Id     int
}

var (
	Sword = Weapon{
		Name:   "Sword",
		Damage: 20,
		Id:     1,
	}
	Axe = Weapon{
		Name:   "Axe",
		Damage: 25,
		Id:     2,
	}
	DoubleAxes = Weapon{
		Name:   "DoubleAxes",
		Damage: 33,
		Id:     3,
	}
	Spear = Weapon{
		Name:   "Spear",
		Damage: 50,
		Id:     4,
	}
)

var AllWeapons = map[string]Weapon{
	"Sword":      Sword,
	"Axe":        Axe,
	"DoubleAxes": DoubleAxes,
	"Spear":      Spear,
}
