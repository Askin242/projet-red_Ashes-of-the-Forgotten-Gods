package structures

type Spells struct {
	Name   string
	Cost   int
	Damage int
}

var (
	Fireball = Spells{
		Name:   "Fireball",
		Damage: 50,
		Cost:   30,
	}
	HandPunch = Spells{
		Name:   "Ice Blast",
		Damage: 30,
		Cost:   20,
	}
)

var AllSpells = map[string]Spells{
	"Fireball":  Fireball,
	"HandPunch": HandPunch,
}
