package structures

type Races struct {
	Name        string
	BonusHP     int
	BonusDamage int
	Skill       Spells
	BonusMana   int
}

var (
	Human = Races{
		Name:        "Human",
		BonusHP:     0,
		BonusMana:   0,
		BonusDamage: 10,
		Skill:       AllSpells["HandPunch"],
	}
	Elf = Races{
		Name:        "Elf",
		BonusHP:     -20,
		BonusMana:   70,
		BonusDamage: 10,
		Skill:       AllSpells["Fireball"],
	}
	Dwarf = Races{
		Name:        "Dwarf",
		BonusHP:     20,
		BonusMana:   30,
		BonusDamage: 10,
		Skill:       AllSpells["Fireball"],
	}
)

var AllRaces = map[string]Races{
	"Human": Human,
	"Elf":   Elf,
	"Dwarf": Dwarf,
}
