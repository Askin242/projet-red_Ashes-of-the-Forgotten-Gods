package ui

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"main/pkg/save"
	"main/pkg/structures"

	"github.com/awesome-gocui/gocui"
)

var (
	blacksmithSelected int
	cachedCraftEntries []craftEntry
)

type craftEntry struct {
	label      string
	kind       string // weapon | armor
	weaponName string
	armorType  string // Helmet | Chestplate | Boots
	armorName  string
}

func getSortedKeys[T any](m map[string]T) []string { // Need to sort to avoid random order
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func addArmorEntries(entries *[]craftEntry, armorType string) {
	var keys []string
	switch armorType {
	case "Helmet":
		keys = getSortedKeys(structures.AllHelmets)
	case "Chestplate":
		keys = getSortedKeys(structures.AllChestplates)
	case "Boots":
		keys = getSortedKeys(structures.AllBoots)
	}

	for _, name := range keys {
		*entries = append(*entries, craftEntry{
			label:     fmt.Sprintf("%s: %s", armorType, name),
			kind:      "armor",
			armorType: armorType,
			armorName: name,
		})
	}
}

func buildCraftEntries() []craftEntry {
	if cachedCraftEntries != nil {
		return cachedCraftEntries
	}

	entries := []craftEntry{}

	weaponKeys := getSortedKeys(structures.AllWeapons)
	for _, name := range weaponKeys {
		entries = append(entries, craftEntry{
			label:      fmt.Sprintf("Weapon: %s", name),
			kind:       "weapon",
			weaponName: name,
		})
	}

	addArmorEntries(&entries, "Helmet")
	addArmorEntries(&entries, "Chestplate")
	addArmorEntries(&entries, "Boots")

	cachedCraftEntries = entries
	return entries
}

func getArmor(armorType, armorName string) structures.Armors {
	switch armorType {
	case "Helmet":
		return structures.AllHelmets[armorName]
	case "Chestplate":
		return structures.AllChestplates[armorName]
	case "Boots":
		return structures.AllBoots[armorName]
	default:
		return structures.Armors{}
	}
}

func computeRequirements(entry craftEntry) (minutes int, materials map[string]int, gold int) {
	switch entry.kind {
	case "weapon":
		w := structures.AllWeapons[entry.weaponName]
		m, req := computeWeaponReq(w)
		return m, req, 5
	case "armor":
		a := getArmor(entry.armorType, entry.armorName)
		m, req, gold := computeArmorReq(a)
		return m, req, gold
	}
	return 0, map[string]int{}, 0
}

func computeWeaponReq(w structures.Weapon) (int, map[string]int) {
	rarity := rarityFromWeaponDamage(w.Damage)
	mats := map[string]int{}
	switch rarity {
	case 4:
		mats["OrcTusk"] = 3
		mats["SkeletonBone"] = 2
		return 10, mats
	case 3:
		mats["SkeletonBone"] = 2
		mats["GoblinEar"] = 1
		return 7, mats
	case 2:
		mats["GoblinEar"] = 2
		return 5, mats
	default:
		mats["GoblinEar"] = 1
		return 3, mats
	}
}

func computeArmorReq(a structures.Armors) (int, map[string]int, int) {
	rarity := rarityFromArmorName(a.Name)
	mats := map[string]int{}
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

func rarityFromWeaponDamage(dmg int) int {
	switch {
	case dmg >= 25:
		return 4
	case dmg >= 18:
		return 3
	case dmg >= 13:
		return 2
	default:
		return 1
	}
}

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

func updateBsHover(g *gocui.Gui) {
	mx, my := g.MousePosition()
	listView, _ := g.View("bs_list")
	if listView == nil {
		return
	}
	x0, y0, x1, y1 := listView.Dimensions()
	if mx >= x0 && mx <= x1 && my >= y0 && my <= y1 {
		idx := my - y0 - 1
		entries := buildCraftEntries()
		if IsValidIndex(idx, len(entries)) && idx != blacksmithSelected {
			blacksmithSelected = idx
			g.Update(func(*gocui.Gui) error { return nil })
		}
	}
}

func attemptCraft(g *gocui.Gui, blacksmith *structures.CraftingBlacksmith, player *structures.Player, entry craftEntry) error {
	if blacksmith.Current != nil {
		return ShowMessageWithOk(g, "bs", "Blacksmith", "Blacksmith is already working on a job", 60, 7)
	}
	var ok bool
	switch entry.kind {
	case "weapon":
		ok = blacksmith.RequestCraftWeapon(player, entry.weaponName)
	case "armor":
		ok = blacksmith.RequestCraftArmor(player, entry.armorType, entry.armorName)
	}
	if ok {
		_ = save.SaveAny("player", player)
		_ = save.SaveAny("blacksmith_job", blacksmith.Current)
		return ShowMessageWithOk(g, "bs", "Blacksmith", "Crafting started!", 60, 7)
	}
	return ShowMessageWithOk(g, "bs", "Blacksmith", "Cannot start crafting (check gold/materials or weight)", 60, 7)
}

func resetBlacksmithCache() {
	cachedCraftEntries = nil
	blacksmithSelected = 0
}

func ShowBlacksmithMenu(blacksmith *structures.CraftingBlacksmith, player *structures.Player) {
	resetBlacksmithCache()
	g, _ := gocui.NewGui(gocui.OutputNormal, false)
	defer g.Close()

	g.SetManagerFunc(func(g *gocui.Gui) error { return blacksmithLayout(g, blacksmith, player) })
	blacksmithKeybindings(g, blacksmith, player)
	g.MainLoop()
}

func blacksmithLayout(g *gocui.Gui, blacksmith *structures.CraftingBlacksmith, player *structures.Player) error {
	maxX, maxY := g.Size()

	updateBsHover(g)

	if err := SetOrUpdateView(g, "bs_title", 0, 0, maxX-1, 2, func(v *gocui.View) {
		v.Frame = false
	}, func(v *gocui.View) {
		fmt.Fprintf(v, "  Blacksmith • Gold: %d\n", player.Money)
	}); err != nil {
		return err
	}

	usableH := maxY - 8
	if usableH < 6 {
		usableH = 6
	}
	topH := usableH / 2
	bottomH := usableH - topH

	listWidth := maxX - 2
	leftWidth := listWidth / 2
	if leftWidth < 30 {
		leftWidth = listWidth / 2
	}

	entries := buildCraftEntries()
	if v, err := g.SetView("bs_list", 1, 3, 1+leftWidth, 3+topH, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Title = " Craftable items "
		v.Highlight = false
	}
	if v, err := g.View("bs_list"); err == nil {
		lines := make([]string, 0, len(entries))
		for _, e := range entries {
			lines = append(lines, e.label)
		}
		RenderListWithHighlight(v, lines, blacksmithSelected)
		if len(entries) == 0 {
			fmt.Fprintln(v, "(Nothing to craft)")
		}
	}

	rightX0 := 1 + leftWidth + 1
	if v, err := g.SetView("bs_side", rightX0, 3, 1+listWidth, 3+topH, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Title = " Details "
	}
	if v, err := g.View("bs_side"); err == nil {
		v.Clear()
		if blacksmith.Current == nil {
			fmt.Fprintln(v, "Status: Idle")
		} else {
			remaining := time.Until(blacksmith.Current.ReadyAt)
			if remaining < 0 {
				remaining = 0
			}
			fmt.Fprintf(v, "Status: Working\n")
			fmt.Fprintf(v, "Job: %s\n", describeJob(blacksmith.Current))
			fmt.Fprintf(v, "Ready in: %s\n\n", remaining.Truncate(time.Second))
		}

		entries := buildCraftEntries()
		if len(entries) == 0 {
			fmt.Fprintln(v, "No craftable items available.")
		} else {
			ValidateSelectedIndex(&blacksmithSelected, len(entries))
			entry := entries[blacksmithSelected]
			mins, mats, gold := computeRequirements(entry)
			fmt.Fprintf(v, "Selected: %s\n", entry.label)
			fmt.Fprintf(v, "Time: %d min\n", mins)
			goldMark := "✗"
			if player.Money >= gold {
				goldMark = "✓"
			}
			fmt.Fprintf(v, "Gold: %d (you: %d) %s\n", gold, player.Money, goldMark)
			if entry.kind == "weapon" {
				w := structures.AllWeapons[entry.weaponName]
				fmt.Fprintf(v, "Damage: %d\n", w.Damage)
				fmt.Fprintf(v, "Defense: 0\n")
			} else {
				a := getArmor(entry.armorType, entry.armorName)
				fmt.Fprintf(v, "Damage: 0\n")
				fmt.Fprintf(v, "Defense: %d\n", a.Defense)
			}
			if len(mats) == 0 {
				fmt.Fprintln(v, "Materials needed: none")
			} else {
				fmt.Fprintln(v, "Materials needed:")
				keys := make([]string, 0, len(mats))
				for k := range mats {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				for _, k := range keys {
					need := mats[k]
					have := player.CountMaterial(k)
					mark := "✗"
					if have >= need {
						mark = "✓"
					}
					fmt.Fprintf(v, "  - %s: need %d, you %d %s\n", k, need, have, mark)
				}
			}
			fmt.Fprintln(v, "")
		}
	}

	invTopY := 3 + topH + 1
	if v, err := g.SetView("bs_build", 1, invTopY, 1+listWidth, invTopY+bottomH, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Title = " Current build "
	}
	if v, err := g.View("bs_build"); err == nil {
		v.Clear()
		wdmg := player.Weapon.Damage
		fmt.Fprintf(v, "Weapon: %s (Damage %d)\n", player.Weapon.Name, wdmg)
		h := player.Entity.Helmet
		c := player.Entity.Chestplate
		b := player.Entity.Boots
		fmt.Fprintf(v, "Helmet: %s (Def %d)\n", h.Name, h.Defense)
		fmt.Fprintf(v, "Chest: %s (Def %d)\n", c.Name, c.Defense)
		fmt.Fprintf(v, "Boots: %s (Def %d)\n", b.Name, b.Defense)
		baseDef := h.Defense + c.Defense + b.Defense
		setBonus := structures.GetSetBonusDefense(player.Entity)
		totalDef := baseDef + setBonus
		if setBonus > 0 {
			fmt.Fprintf(v, "Set bonus: +%d\n", setBonus)
		}
		fmt.Fprintf(v, "Total defense: %d\n", totalDef)
	}

	closeBtnX := maxX - 12
	if closeBtnX < 2 {
		closeBtnX = 2
	}
	closeBtnY := maxY - 3
	if closeBtnY < 3 {
		closeBtnY = 3
	}
	craftBtnX := closeBtnX - 12
	if craftBtnX < 2 {
		craftBtnX = 2
	}
	createButton(g, "bs_craft", " Craft ", craftBtnX, closeBtnY, 10, 2, "bs_craft")
	createButton(g, "bs_close", " Close ", closeBtnX, closeBtnY, 10, 2, "bs_close")

	if blacksmith.Current != nil && time.Now().After(blacksmith.Current.ReadyAt) {
		collectX := craftBtnX - 14
		if collectX < 2 {
			collectX = 2
		}
		createButton(g, "bs_collect", " Collect ", collectX, closeBtnY, 12, 2, "bs_collect")
	} else {
		g.DeleteView("bs_collect")
	}

	g.SetCurrentView("bs_list")
	return nil
}

func describeJob(job *structures.CraftJob) string {
	if job == nil {
		return "None"
	}
	switch job.Request.OutputType {
	case "weapon":
		return fmt.Sprintf("Weapon: %s", job.Request.WeaponName)
	case "armor":
		return fmt.Sprintf("%s: %s", job.Request.ArmorType, job.Request.ArmorName)
	default:
		return "Unknown"
	}
}

func blacksmithKeybindings(g *gocui.Gui, blacksmith *structures.CraftingBlacksmith, player *structures.Player) error {
	BindQuitOnEsc(g)

	BindListNavigation(g, "bs_list", &blacksmithSelected, func() int { return len(buildCraftEntries()) })

	g.SetKeybinding("bs_list", gocui.KeyEnter, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		entries := buildCraftEntries()
		if len(entries) == 0 || !IsValidIndex(blacksmithSelected, len(entries)) {
			return nil
		}
		entry := entries[blacksmithSelected]
		return attemptCraft(g, blacksmith, player, entry)
	})

	g.SetKeybinding("", 'c', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if blacksmith.Current == nil {
			return nil
		}
		if time.Now().Before(blacksmith.Current.ReadyAt) {
			return ShowMessageWithOk(g, "bs", "Blacksmith", "Not ready yet", 60, 7)
		}
		return showConfirmCollect(g, blacksmith, player)
	})

	EnableMouseAndSetHandler(g, func(g *gocui.Gui, v *gocui.View) error {
		mx, my := g.MousePosition()

		if listView, _ := g.View("bs_list"); listView != nil {
			x0, y0, x1, y1 := listView.Dimensions()
			if mx >= x0 && mx <= x1 && my >= y0 && my <= y1 {
				idx := my - y0 - 1
				entries := buildCraftEntries()
				if IsValidIndex(idx, len(entries)) {
					blacksmithSelected = idx
					g.Update(func(*gocui.Gui) error { return nil })
					entry := entries[idx]
					return attemptCraft(g, blacksmith, player, entry)
				}
			}
		}

		buttons := []ButtonHandler{
			{"bs_ok", func(g *gocui.Gui, v *gocui.View) error {
				DeleteViews(g, "bs_msg", "bs_ok")
				return nil
			}},
			{"bs_confirm_yes", func(g *gocui.Gui, v *gocui.View) error {
				return performReplaceCollect(g, blacksmith, player)
			}},
			{"bs_confirm_no", func(g *gocui.Gui, v *gocui.View) error {
				DeleteViews(g, "bs_confirm", "bs_confirm_yes", "bs_confirm_no")
				return nil
			}},
			{"bs_close", func(g *gocui.Gui, v *gocui.View) error {
				return gocui.ErrQuit
			}},
			{"bs_craft", func(g *gocui.Gui, v *gocui.View) error {
				entries := buildCraftEntries()
				if len(entries) == 0 || !IsValidIndex(blacksmithSelected, len(entries)) {
					return nil
				}
				entry := entries[blacksmithSelected]
				return attemptCraft(g, blacksmith, player, entry)
			}},
			{"bs_collect", func(g *gocui.Gui, v *gocui.View) error {
				if blacksmith.Current == nil {
					return nil
				}
				if time.Now().Before(blacksmith.Current.ReadyAt) {
					return ShowMessageWithOk(g, "bs", "Blacksmith", "Not ready yet", 60, 7)
				}
				return showConfirmCollect(g, blacksmith, player)
			}},
		}
		if err := HandleMouseClickButtons(g, mx, my, buttons); err != nil {
			return err
		}
		return nil
	})
	return nil
}

func showConfirmCollect(g *gocui.Gui, blacksmith *structures.CraftingBlacksmith, player *structures.Player) error {
	maxX, maxY := g.Size()
	w := 64
	h := 9
	x := (maxX - w) / 2
	y := (maxY - h) / 2

	msg := "Collecting will replace your currently equipped item and erase it. Proceed?"
	if v, err := g.SetView("bs_confirm", x, y, x+w, y+h, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Title = " Confirm Replace "
		fmt.Fprintf(v, "\n  %s\n\n", msg)
	}
	yesX := x + w - 28
	noX := x + w - 14
	btnY := y + h - 2
	createButton(g, "bs_confirm_yes", " Yes ", yesX, btnY-1, 10, 2, "bs_confirm_yes")
	createButton(g, "bs_confirm_no", " No ", noX, btnY-1, 10, 2, "bs_confirm_no")

	g.SetKeybinding("bs_confirm_yes", gocui.KeyEnter, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return performReplaceCollect(g, blacksmith, player)
	})
	g.SetKeybinding("bs_confirm", gocui.KeyEsc, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		DeleteViews(g, "bs_confirm", "bs_confirm_yes", "bs_confirm_no")
		return nil
	})
	return nil
}

func performReplaceCollect(g *gocui.Gui, blacksmith *structures.CraftingBlacksmith, player *structures.Player) error {
	if blacksmith.Current == nil {
		return nil
	}
	if time.Now().Before(blacksmith.Current.ReadyAt) {
		return ShowMessageWithOk(g, "bs", "Blacksmith", "Not ready yet", 60, 7)
	}
	switch blacksmith.Current.Request.OutputType {
	case "weapon":
		w := structures.AllWeapons[blacksmith.Current.Request.WeaponName]
		player.Weapon = w
	case "armor":
		a := getArmor(blacksmith.Current.Request.ArmorType, blacksmith.Current.Request.ArmorName)
		switch blacksmith.Current.Request.ArmorType {
		case "Helmet":
			player.Entity.Helmet = a
		case "Chestplate":
			player.Entity.Chestplate = a
		case "Boots":
			player.Entity.Boots = a
		}
	}
	blacksmith.Current = nil
	_ = save.SaveAny("player", player)
	_ = save.SaveAny("blacksmith_job", blacksmith.Current)

	DeleteViews(g, "bs_confirm", "bs_confirm_yes", "bs_confirm_no")
	return ShowMessageWithOk(g, "bs", "Blacksmith", "Equipped new item from the blacksmith!", 60, 7)
}
