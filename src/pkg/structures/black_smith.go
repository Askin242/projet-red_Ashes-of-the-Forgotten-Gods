package structures

import "time"

type BlackSmith struct {
	Entity
	Inventory       Inventory
	FirstHealBought bool
}

type CraftRequest struct {
	OutputType string // "weapon" | "armor"
	WeaponName string // if OutputType==weapon
	ArmorType  string // Helmet/Chestplate/Boots if OutputType==armor
	ArmorName  string // StormBringer/SunBreaker/VoidWalker
}

type CraftJob struct {
	Request   CraftRequest
	ReadyAt   time.Time
	GoldCost  int
	Materials map[string]int
}

func craftingRulesForWeapon(w Weapon) (minutes int, mats map[string]int) {
	rarity := rarityFromWeaponDamage(w.Damage)
	mats = map[string]int{}
	switch rarity {
	case 4:
		mats["OrcTusk"] = 3
		mats["SkeletonBone"] = 2
		return 3, mats
	case 3:
		mats["SkeletonBone"] = 2
		mats["GoblinEar"] = 1
		return 2, mats
	case 2:
		mats["GoblinEar"] = 2
		return 1, mats
	default:
		mats["GoblinEar"] = 1
		return 1, mats
	}
}

func craftingRulesForArmor(a Armors) (minutes int, mats map[string]int, gold int) {
	rarity := rarityFromArmorName(a.Name)
	mats = map[string]int{}
	baseGold := 5
	typeMult := 1
	switch a.Type {
	case "Chestplate":
		typeMult = 2
		baseGold = 8
	case "Helmet", "Boots":
		typeMult = 1
		baseGold = 5
	}
	switch rarity {
	case 6:
		mats["OrcTusk"] = 4 * typeMult
		mats["SkeletonBone"] = 2 * typeMult
		return 10, mats, baseGold
	case 4:
		mats["OrcTusk"] = 2 * typeMult
		mats["SkeletonBone"] = 2 * typeMult
		return 7, mats, baseGold
	default:
		mats["GoblinEar"] = 2 * typeMult
		return 3, mats, baseGold
	}
}

type CraftingBlacksmith struct {
	BlackSmith
	Current *CraftJob
}

func (cb *CraftingBlacksmith) RequestCraftWeapon(player *Player, weaponName string) bool {
	if cb.Current != nil {
		return false
	}
	w, ok := AllWeapons[weaponName]
	if !ok {
		return false
	}
	if player.Money < 5 {
		return false
	}
	minutes, mats := craftingRulesForWeapon(w)
	if !player.HasMaterialsBatch(mats) {
		return false
	}
	player.Money -= 5
	if !player.RemoveMaterialsBatch(mats) {
		return false
	}

	job := CraftJob{
		Request:   CraftRequest{OutputType: "weapon", WeaponName: weaponName},
		ReadyAt:   time.Now().Add(time.Duration(minutes) * time.Minute),
		GoldCost:  5,
		Materials: mats,
	}
	cb.Current = &job
	return true
}

func (cb *CraftingBlacksmith) RequestCraftArmor(player *Player, armorType, armorName string) bool {
	if cb.Current != nil {
		return false
	}
	var a Armors
	switch armorType {
	case "Helmet":
		a = AllHelmets[armorName]
	case "Chestplate":
		a = AllChestplates[armorName]
	case "Boots":
		a = AllBoots[armorName]
	default:
		return false
	}
	if a.Name == "" {
		return false
	}
	if player.Money < 5 {
		return false
	}
	minutes, mats, gold := craftingRulesForArmor(a)
	if !player.HasMaterialsBatch(mats) {
		return false
	}
	if player.Money < gold {
		return false
	}
	player.Money -= gold
	if !player.RemoveMaterialsBatch(mats) {
		return false
	}

	job := CraftJob{
		Request:   CraftRequest{OutputType: "armor", ArmorType: armorType, ArmorName: armorName},
		ReadyAt:   time.Now().Add(time.Duration(minutes) * time.Minute),
		GoldCost:  gold,
		Materials: mats,
	}
	cb.Current = &job
	return true
}

func (cb *CraftingBlacksmith) CollectReady(player *Player) int {
	if cb.Current == nil {
		return 0
	}
	if time.Now().Before(cb.Current.ReadyAt) {
		return 0
	}
	switch cb.Current.Request.OutputType {
	case "weapon":
		w := AllWeapons[cb.Current.Request.WeaponName]
		if player.AddItem(NewWeaponItem(w)) {
			cb.Current = nil
			return 1
		}
	case "armor":
		var a Armors
		switch cb.Current.Request.ArmorType {
		case "Helmet":
			a = AllHelmets[cb.Current.Request.ArmorName]
		case "Chestplate":
			a = AllChestplates[cb.Current.Request.ArmorName]
		case "Boots":
			a = AllBoots[cb.Current.Request.ArmorName]
		}
		if player.AddItem(NewArmorItem(a)) {
			cb.Current = nil
			return 1
		}
	}
	return 0
}
