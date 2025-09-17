package structures

import (
	"math/rand"
	"time"
)

type Item struct {
	Name   string
	Id     int
	Weight int
	Price  int
	Rarity int
}

type InventoryEntry interface {
	GetItem() Item
}

func (it Item) GetItem() Item { return it }

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

func newItemID() int {
	return 1 + rng.Intn(2000000000)
}

func NewItem(name string, weight int, price int, rarity int) Item {
	return Item{
		Name:   name,
		Id:     newItemID(),
		Weight: weight,
		Price:  price,
		Rarity: rarity,
	}
}

type Material struct {
	Item
	Key string
}

func (m Material) GetItem() Item { return m.Item }

func NewMaterial(key string, displayName string) Material {
	return Material{
		Item: NewItem(displayName, 0, 0, 0),
		Key:  key,
	}
}

var (
	GoblinEar    = NewMaterial("GoblinEar", "Goblin Ear")
	SkeletonBone = NewMaterial("SkeletonBone", "Skeleton Bone")
	OrcTusk      = NewMaterial("OrcTusk", "Orc Tusk")
)

var AllMaterials = map[string]Material{
	"GoblinEar":    GoblinEar,
	"SkeletonBone": SkeletonBone,
	"OrcTusk":      OrcTusk,
}

func GenerateLootFromEnemy(r EnemyRace) InventoryEntry {
	switch r.Name {
	case "Goblin":
		return GoblinEar
	case "Skeleton":
		return SkeletonBone
	case "Orc":
		return OrcTusk
	default:
		return GoblinEar
	}
}

type Potion struct {
	Item
	Size int
	Type string
}

func (p Potion) GetItem() Item { return p.Item }

func GetPotion(potionType string, size int, price int) Potion {
	return Potion{
		Size: size,
		Type: potionType,
		Item: NewItem(potionType+" Potion", 1, price, 1),
	}
}

var (
	Poison = Potion{
		Size: 1,
		Type: "Poison",
		Item: NewItem("Poison Potion", 1, 80, 2),
	}
	Cure = Potion{
		Size: 1,
		Type: "Cure",
		Item: NewItem("Cure Potion", 1, 80, 4),
	}
	Heal = Potion{
		Size: 1,
		Type: "Heal",
		Item: NewItem("Heal Potion", 1, 50, 1),
	}
)

var AllPotions = map[string]Potion{
	"Poison": Poison,
	"Cure":   Cure,
	"Heal":   Heal,
}

type Spellbooks struct {
	Item
	Spell Spell
}

func (s Spellbooks) GetItem() Item { return s.Item }

var (
	SpellBookFireball = Spellbooks{
		Item:  NewItem("Fireball Spellbook", 1, 200, 4),
		Spell: AllSpells["Fireball"],
	}
)

var AllSpellbooks = map[string]Spellbooks{
	"SpellBookFireball": SpellBookFireball,
}

func weightFromRarity(rarity int) int {
	weight := 7 - rarity
	if weight < 1 {
		weight = 1
	}
	return weight
}

func GetRandomItemByRarity() InventoryEntry { // Chatgpt based ♥
	type weightedEntry struct {
		entry  InventoryEntry
		weight int
	}

	var pool []weightedEntry

	for _, p := range AllPotions {
		pool = append(pool, weightedEntry{entry: p, weight: weightFromRarity(p.Item.Rarity)})
	}

	for _, sb := range AllSpellbooks {
		pool = append(pool, weightedEntry{entry: sb, weight: weightFromRarity(sb.Item.Rarity)})
	}

	for _, w := range AllWeapons {
		wi := NewWeaponItem(w)
		pool = append(pool, weightedEntry{entry: wi, weight: weightFromRarity(wi.Item.Rarity)})
	}

	for _, a := range AllHelmets {
		ai := NewArmorItem(a)
		pool = append(pool, weightedEntry{entry: ai, weight: weightFromRarity(ai.Item.Rarity)})
	}
	for _, a := range AllChestplates {
		ai := NewArmorItem(a)
		pool = append(pool, weightedEntry{entry: ai, weight: weightFromRarity(ai.Item.Rarity)})
	}
	for _, a := range AllBoots {
		ai := NewArmorItem(a)
		pool = append(pool, weightedEntry{entry: ai, weight: weightFromRarity(ai.Item.Rarity)})
	}

	if len(pool) == 0 {
		return Heal
	}

	total := 0
	for _, we := range pool {
		total += we.weight
	}
	if total <= 0 {
		return Heal
	}

	r := rng.Intn(total)
	cumulative := 0
	for _, we := range pool {
		cumulative += we.weight
		if r < cumulative {
			return we.entry
		}
	}
	return pool[len(pool)-1].entry
}

type WeaponItem struct {
	Item
	Weapon Weapon
}

func (w WeaponItem) GetItem() Item { return w.Item }

func rarityFromWeaponDamage(damage int) int {
	switch {
	case damage >= 25:
		return 4
	case damage >= 18:
		return 3
	case damage >= 13:
		return 2
	default:
		return 1
	}
}

func NewWeaponItem(weapon Weapon) WeaponItem {
	return WeaponItem{
		Item:   NewItem(weapon.Name, 0, weapon.Damage*10, rarityFromWeaponDamage(weapon.Damage)),
		Weapon: weapon,
	}
}

type ArmorItem struct {
	Item
	Armor Armors
}

func (a ArmorItem) GetItem() Item { return a.Item }

func rarityFromArmorName(name string) int {
	switch name {
	case "StormBringer":
		return 6
	case "SunBreaker":
		return 4
	default:
		return 2
	}
}

func NewArmorItem(armor Armors) ArmorItem {
	label := armor.Type + " " + armor.Name
	return ArmorItem{
		Item:  NewItem(label, 0, armor.Defense*10, rarityFromArmorName(armor.Name)),
		Armor: armor,
	}
}
