package structures

type Item struct {
	Name   string
	Weight int
}

type InventoryEntry interface {
	GetItem() Item
}

func (it Item) GetItem() Item { return it }

type Potion struct {
	Item
	Size int
	Type string
}

func (p Potion) GetItem() Item { return p.Item }

var (
	Poison = Potion{
		Size: 1,
		Type: "Poison",
		Item: Item{
			Name:   "Poison Potion",
			Weight: 1,
		},
	}
	Cure = Potion{
		Size: 1,
		Type: "Cure",
		Item: Item{
			Name:   "Cure Potion",
			Weight: 1,
		},
	}
)

var AllPotions = map[string]Potion{
	"Poison": Poison,
	"Cure":   Cure,
}

type Spellbooks struct {
	Item
	Spell string
}

func (s Spellbooks) GetItem() Item { return s.Item }

var (
	SpellBookFireball = Spellbooks{
		Item: Item{
			Name:   "Fireball",
			Weight: 1,
		},
		Spell: "Fireball",
	}
)

var AllSpellbooks = map[string]Spellbooks{
	"SpellBookFireball": SpellBookFireball,
}
