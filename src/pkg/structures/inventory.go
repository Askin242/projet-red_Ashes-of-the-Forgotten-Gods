package structures

import "encoding/json"

type Inventory []InventoryEntry

func (inv *Inventory) UnmarshalJSON(data []byte) error { // Required for json.Unmarshal to work with Inventory else it does not parse correctly
	var rawSlice []map[string]any
	if err := json.Unmarshal(data, &rawSlice); err != nil {
		return err
	}

	var entries []InventoryEntry
	for _, m := range rawSlice {
		b, err := json.Marshal(m)
		if err != nil {
			return err
		}

		switch {
		case m["Key"] != nil:
			var mat Material
			if err := json.Unmarshal(b, &mat); err != nil {
				return err
			}
			entries = append(entries, mat)
		case m["Spell"] != nil:
			var sb Spellbooks
			if err := json.Unmarshal(b, &sb); err != nil {
				return err
			}
			entries = append(entries, sb)
		case m["Armor"] != nil:
			var ai ArmorItem
			if err := json.Unmarshal(b, &ai); err != nil {
				return err
			}
			entries = append(entries, ai)
		case m["Weapon"] != nil:
			var wi WeaponItem
			if err := json.Unmarshal(b, &wi); err != nil {
				return err
			}
			entries = append(entries, wi)
		case m["Size"] != nil && m["Type"] != nil:
			var p Potion
			if err := json.Unmarshal(b, &p); err != nil {
				return err
			}
			entries = append(entries, p)
		default:
			var it Item
			if err := json.Unmarshal(b, &it); err != nil {
				return err
			}
			entries = append(entries, it)
		}
	}

	*inv = entries
	return nil
}
