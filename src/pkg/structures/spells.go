package structures

type Spell struct {
	Name   string
	Cost   int
	Damage int
}

var (
	Fireball = Spell{
		Name:   "Fireball",
		Damage: 50,
		Cost:   30,
	}
	HandPunch = Spell{
		Name:   "Ice Blast",
		Damage: 30,
		Cost:   20,
	}
	None = Spell{
		Name:   "None",
		Damage: 0,
		Cost:   0,
	}
)

var AllSpells = map[string]Spell{
	"Fireball":  Fireball,
	"HandPunch": HandPunch,
	"None":      None,
}
