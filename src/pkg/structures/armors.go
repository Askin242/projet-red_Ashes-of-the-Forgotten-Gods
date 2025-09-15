package structures

import (
	"math/rand"
)

type Armors struct {
	Name    string
	Type    string
	Defense int
}

var (
	HelmetStormBringer     = Armors{Name: "StormBringer", Type: "Helmet", Defense: 6}
	HelmetSunBreaker       = Armors{Name: "SunBreaker", Type: "Helmet", Defense: 4}
	HelmetVoidWalker       = Armors{Name: "VoidWalker", Type: "Helmet", Defense: 2}
	ChestplateStormBringer = Armors{Name: "StormBringer", Type: "Chestplate", Defense: 12}
	ChestplateSunBreaker   = Armors{Name: "SunBreaker", Type: "Chestplate", Defense: 8}
	ChestplateVoidWalker   = Armors{Name: "VoidWalker", Type: "Chestplate", Defense: 4}
	BootsStormBringer      = Armors{Name: "StormBringer", Type: "Boots", Defense: 6}
	BootsSunBreaker        = Armors{Name: "SunBreaker", Type: "Boots", Defense: 4}
	BootsVoidWalker        = Armors{Name: "VoidWalker", Type: "Boots", Defense: 2}
)

var setBonusDefense = map[string]int{
	"StormBringer": 6,
	"SunBreaker":   4,
	"VoidWalker":   2,
}

var armorRarityWeight = map[string]int{
	"StormBringer": 1,
	"SunBreaker":   3,
	"VoidWalker":   6,
}

var AllHelmets = map[string]Armors{
	"StormBringer": HelmetStormBringer,
	"SunBreaker":   HelmetSunBreaker,
	"VoidWalker":   HelmetVoidWalker,
}

var AllChestplates = map[string]Armors{
	"StormBringer": ChestplateStormBringer,
	"SunBreaker":   ChestplateSunBreaker,
	"VoidWalker":   ChestplateVoidWalker,
}

var AllBoots = map[string]Armors{
	"StormBringer": BootsStormBringer,
	"SunBreaker":   BootsSunBreaker,
	"VoidWalker":   BootsVoidWalker,
}

func GetSetBonusDefense(ent Entity) int {
	if ent.Helmet.Name == ent.Chestplate.Name && ent.Chestplate.Name == ent.Boots.Name && ent.Helmet.Name != "None" {
		if bonus, ok := setBonusDefense[ent.Helmet.Name]; ok {
			return bonus
		}
	}
	return 0
}

func getWeightedRandomName() string { // Get random armor name based on rarity
	total := 0
	for _, w := range armorRarityWeight {
		total += w
	}
	if total <= 0 {
		return "VoidWalker"
	}
	r := rand.Intn(total)
	cumulative := 0
	for name, w := range armorRarityWeight {
		cumulative += w
		if r < cumulative {
			return name
		}
	}
	return "VoidWalker"
}

func GetRandomArmorByType(armorType string) Armors {
	name := getWeightedRandomName()
	switch armorType {
	case "Helmet":
		return AllHelmets[name]
	case "Chestplate":
		return AllChestplates[name]
	case "Boots":
		return AllBoots[name]
	default:
		return Armors{Name: "None", Type: armorType, Defense: 0}
	}
}
