package structures

type Spell struct {
	Name    string
	Cost    int
	Damage  int
	Element string
}

var (
	Fireball = Spell{
		Name:    "Fireball",
		Damage:  9,
		Cost:    30,
		Element: "Fire",
	}
	PoisonFlask = Spell{
		Name:    "Poison Flask",
		Damage:  5,
		Cost:    15,
		Element: "Poison",
	}
	LightningStrike = Spell{
		Name:    "Lightning Strike",
		Damage:  12,
		Cost:    40,
		Element: "Lightning",
	}
	IceBlast = Spell{
		Name:    "Ice Blast",
		Damage:  12,
		Cost:    40,
		Element: "Ice",
	}
	HandPunch = Spell{
		Name:    "Hand Punch",
		Damage:  7,
		Cost:    20,
		Element: "Neutral",
	}
	None = Spell{
		Name:    "None",
		Damage:  0,
		Cost:    0,
		Element: "Neutral",
	}
)

var AllSpells = map[string]Spell{
	"Fireball":        Fireball,
	"HandPunch":       HandPunch,
	"PoisonFlask":     PoisonFlask,
	"LightningStrike": LightningStrike,
	"IceBlast":        IceBlast,
	"None":            None,
}
