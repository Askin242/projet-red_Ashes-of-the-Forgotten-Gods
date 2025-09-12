package structures

type Potion struct {
	Size int
	Type string
	Name string
}

var (
	Poison = Potion{
		Size: 1,
		Type: "Poison",
		Name: "Poison Potion",
	}
	Cure = Potion{
		Size: 1,
		Type: "Cure",
		Name: "Cure Potion",
	}
)

var AllPotions = map[string]Potion{
	"Poison": Poison,
	"Cure":   Cure,
}
