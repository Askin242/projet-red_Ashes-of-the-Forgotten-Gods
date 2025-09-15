package structures

type Race struct {
	Name        string
	BonusHP     int
	BonusDamage int
	Skill       Spell
	BonusMana   int
}

type EnemyRace struct {
	Name        string
	BonusHP     int
	BonusDamage int
	Skill       Spell
	BonusMana   int
}

var (
	Human = Race{
		Name:        "Human",
		BonusHP:     0,
		BonusMana:   0,
		BonusDamage: 10,
		Skill:       AllSpells["HandPunch"],
	}
	Elf = Race{
		Name:        "Elf",
		BonusHP:     -20,
		BonusMana:   70,
		BonusDamage: 10,
		Skill:       AllSpells["HandPunch"],
	}
	Dwarf = Race{
		Name:        "Dwarf",
		BonusHP:     20,
		BonusMana:   30,
		BonusDamage: 10,
		Skill:       AllSpells["HandPunch"],
	}
	Orc = EnemyRace{
		Name:        "Orc",
		BonusHP:     10,
		BonusDamage: 15,
	}
	Skeleton = EnemyRace{
		Name:        "Skeleton",
		BonusHP:     -30,
		BonusDamage: 20,
	}
	Goblin = EnemyRace{
		Name:        "Goblin",
		BonusHP:     10,
		BonusDamage: 10,
	}
)

var AllRaces = map[string]Race{
	"Human": Human,
	"Elf":   Elf,
	"Dwarf": Dwarf,
}

var AllEnemyRaces = map[string]EnemyRace{
	"Orc":      Orc,
	"Skeleton": Skeleton,
	"Goblin":   Goblin,
}
