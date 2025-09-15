package structures

type Spell struct {
	Name   string
	Cost   int
	Damage int
}

var (
	Fireball = Spell{
		Name:   "Fireball",
		Damage: 9,
		Cost:   30,
	}
	PoisonFlask = Spell{
		Name:   "Poison Flask",
		Damage: 5,
		Cost:   15,
	}
	LightningStrike = Spell{
		Name:   "Lightning Strike",
		Damage: 12,
		Cost:   40,
	}
	HandPunch = Spell{
		Name:   "Ice Blast",
		Damage: 7,
		Cost:   20,
	}
	None = Spell{
		Name:   "None",
		Damage: 0,
		Cost:   0,
	}
)

var AllSpells = map[string]Spell{
	"Fireball":        Fireball,
	"HandPunch":       HandPunch,
	"PoisonFlask":     PoisonFlask,
	"LightningStrike": LightningStrike,
	"None":            None,
}
