package ui

import (
	"fmt"
	"strings"

	"main/pkg/structures"

	"github.com/awesome-gocui/gocui"
)

var (
	inventoryOpen     = false
	inventorySelected = 0
)

func ShowInventory(g *gocui.Gui, player *structures.Player) error {
	if inventoryOpen {
		return CloseInventory(g)
	}

	inventoryOpen = true
	inventorySelected = 0

	maxX, maxY := g.Size()
	width := 60
	height := 25
	x := (maxX - width) / 2
	y := (maxY - height) / 2

	if v, err := g.SetView("inventory", x, y, x+width, y+height, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = " Inventory (E to close) "
		v.Frame = true
		v.Highlight = false // We'll handle highlighting manually for better control
		updateInventoryView(v, player)
	}

	if _, err := g.SetCurrentView("inventory"); err != nil {
		return err
	}

	return setupInventoryKeybindings(g, player)
}

func CloseInventory(g *gocui.Gui) error {
	inventoryOpen = false
	g.DeleteView("inventory")
	g.DeleteKeybinding("inventory", gocui.KeyArrowUp, gocui.ModNone)
	g.DeleteKeybinding("inventory", gocui.KeyArrowDown, gocui.ModNone)
	g.DeleteKeybinding("inventory", 'e', gocui.ModNone)
	g.DeleteKeybinding("inventory", 'E', gocui.ModNone)
	g.DeleteKeybinding("inventory", gocui.KeyEsc, gocui.ModNone)
	g.DeleteKeybinding("inventory", gocui.KeyEnter, gocui.ModNone)
	return nil
}

func updateInventoryView(v *gocui.View, player *structures.Player) {
	v.Clear()

	fmt.Fprintf(v, " Player: %s\n", player.Entity.Name)
	fmt.Fprintf(v, " Money: %d coins\n", player.Money)
	fmt.Fprintf(v, " Carry Weight: %d/%d\n", player.CurrentCarryWeight(), player.MaxCarryWeight)
	fmt.Fprintln(v, strings.Repeat("-", 56))

	if len(player.Inventory) == 0 {
		fmt.Fprintln(v, " Your inventory is empty.")
		return
	}

	fmt.Fprintln(v, " Items:")
	for i, entry := range player.Inventory {
		item := entry.GetItem()

		var line string
		switch e := entry.(type) {
		case structures.Material:
			line = fmt.Sprintf("[Material] %s", item.Name)
		case structures.Potion:
			line = fmt.Sprintf("[Potion] %s (Size: %d)", item.Name, e.Size)
		case structures.Spellbooks:
			line = fmt.Sprintf("[Spellbook] %s (Spell: %s)", item.Name, e.Spell.Name)
		case structures.WeaponItem:
			line = fmt.Sprintf("[Weapon] %s (Damage: %d)", item.Name, e.Weapon.Damage)
		case structures.ArmorItem:
			line = fmt.Sprintf("[Armor] %s (Defense: %d)", item.Name, e.Armor.Defense)
		default:
			line = fmt.Sprintf("%s (Weight: %d)", item.Name, item.Weight)
		}

		// Add enhanced visual highlighting for selected item
		if i == inventorySelected {
			fmt.Fprintf(v, " \033[43m\033[30m► %s \033[0m\n", line) // Yellow background, black text with arrow
		} else {
			fmt.Fprintf(v, "   %s\n", line)
		}
	}

	fmt.Fprintln(v, strings.Repeat("-", 56))
	fmt.Fprintln(v, " Controls:")
	fmt.Fprintln(v, " ↑/↓ - Navigate  |  Enter - Use Item  |  E/Esc - Close")
}

func setupInventoryKeybindings(g *gocui.Gui, player *structures.Player) error {
	if err := g.SetKeybinding("inventory", gocui.KeyArrowUp, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if len(player.Inventory) > 0 && inventorySelected > 0 {
			inventorySelected--
			updateInventoryView(v, player)
		}
		return nil
	}); err != nil {
		return err
	}

	if err := g.SetKeybinding("inventory", gocui.KeyArrowDown, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if len(player.Inventory) > 0 && inventorySelected < len(player.Inventory)-1 {
			inventorySelected++
			updateInventoryView(v, player)
		}
		return nil
	}); err != nil {
		return err
	}

	if err := g.SetKeybinding("inventory", 'e', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return CloseInventory(g)
	}); err != nil {
		return err
	}

	if err := g.SetKeybinding("inventory", 'E', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return CloseInventory(g)
	}); err != nil {
		return err
	}

	if err := g.SetKeybinding("inventory", gocui.KeyEsc, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return CloseInventory(g)
	}); err != nil {
		return err
	}

	// Use item
	if err := g.SetKeybinding("inventory", gocui.KeyEnter, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return useSelectedItem(g, v, player)
	}); err != nil {
		return err
	}

	return nil
}

func useSelectedItem(g *gocui.Gui, v *gocui.View, player *structures.Player) error {
	if len(player.Inventory) == 0 || inventorySelected >= len(player.Inventory) {
		return nil
	}

	selectedItem := player.Inventory[inventorySelected]

	switch item := selectedItem.(type) {
	case structures.Potion:
		if item.Type == "Heal" {
			if player.UsePotion(item) {
				updateInventoryView(v, player)
				ShowMessageWithOk(g, "potion", "Item Used",
					fmt.Sprintf("Used %s! Restored health.", item.Item.Name), 40, 8)
			}
		}
	case structures.Spellbooks:
		hasSpell := false
		for _, spell := range player.Spells {
			if spell.Name == item.Spell.Name {
				hasSpell = true
				break
			}
		}
		if !hasSpell {
			player.Spells = append(player.Spells, item.Spell)
			player.RemoveItem(selectedItem)
			ensureValidSelection(player)
			updateInventoryView(v, player)
			ShowMessageWithOk(g, "spell", "Spell Learned",
				fmt.Sprintf("Learned spell: %s!", item.Spell.Name), 40, 8)
		} else {
			ShowMessageWithOk(g, "spell", "Already Known",
				"You already know this spell!", 40, 8)
		}
	case structures.WeaponItem:
		player.Weapon = item.Weapon
		player.RemoveItem(selectedItem)
		ensureValidSelection(player)
		updateInventoryView(v, player)
		ShowMessageWithOk(g, "weapon", "Weapon Equipped",
			fmt.Sprintf("Equipped %s!", item.Weapon.Name), 40, 8)
	case structures.ArmorItem:
		switch item.Armor.Type {
		case "Helmet":
			player.Entity.Helmet = item.Armor
		case "Chestplate":
			player.Entity.Chestplate = item.Armor
		case "Boots":
			player.Entity.Boots = item.Armor
		}
		player.RemoveItem(selectedItem)
		ensureValidSelection(player)
		updateInventoryView(v, player)
		ShowMessageWithOk(g, "armor", "Armor Equipped",
			fmt.Sprintf("Equipped %s!", item.Item.Name), 40, 8)
	default:
		showInventoryPopup(g, "Item Info",
			fmt.Sprintf("%s - This item cannot be used directly.", selectedItem.GetItem().Name), player)
	}

	return nil
}

func ensureValidSelection(player *structures.Player) {
	if len(player.Inventory) == 0 {
		inventorySelected = 0
	} else if inventorySelected >= len(player.Inventory) {
		inventorySelected = len(player.Inventory) - 1
	}
}

func disableInventoryKeybindings(g *gocui.Gui) {
	g.DeleteKeybinding("inventory", gocui.KeyArrowUp, gocui.ModNone)
	g.DeleteKeybinding("inventory", gocui.KeyArrowDown, gocui.ModNone)
	g.DeleteKeybinding("inventory", 'e', gocui.ModNone)
	g.DeleteKeybinding("inventory", 'E', gocui.ModNone)
	g.DeleteKeybinding("inventory", gocui.KeyEsc, gocui.ModNone)
	g.DeleteKeybinding("inventory", gocui.KeyEnter, gocui.ModNone)
}

func showInventoryPopup(g *gocui.Gui, title, message string, player *structures.Player) error {
	maxX, maxY := g.Size()
	width := 50
	height := 8
	x := (maxX - width) / 2
	y := (maxY - height) / 2

	msgId := "inventory_popup_msg"
	okId := "inventory_popup_ok"

	if v, err := g.SetView(msgId, x, y, x+width, y+height, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = " " + title + " "
		v.Frame = true
		fmt.Fprintf(v, "\n  %s\n\n", message)
	}

	btnX := x + width - 14
	btnY := y + height - 2
	if v, err := g.SetView(okId, btnX, btnY-1, btnX+10, btnY+1, 0); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = true
		v.BgColor = gocui.ColorYellow
		v.FgColor = gocui.ColorBlack
		fmt.Fprint(v, " OK ")
	}

	g.SetCurrentView(okId)

	g.SetKeybinding(okId, gocui.KeyEnter, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		g.DeleteView(msgId)
		g.DeleteView(okId)
		g.DeleteKeybinding(okId, gocui.KeyEnter, gocui.ModNone)
		g.DeleteKeybinding(okId, gocui.KeyEsc, gocui.ModNone)
		g.DeleteKeybinding(msgId, gocui.KeyEsc, gocui.ModNone)
		g.SetCurrentView("inventory")
		return nil
	})

	g.SetKeybinding(okId, gocui.KeyEsc, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		g.DeleteView(msgId)
		g.DeleteView(okId)
		g.DeleteKeybinding(okId, gocui.KeyEnter, gocui.ModNone)
		g.DeleteKeybinding(okId, gocui.KeyEsc, gocui.ModNone)
		g.DeleteKeybinding(msgId, gocui.KeyEsc, gocui.ModNone)
		g.SetCurrentView("inventory")
		return nil
	})

	g.SetKeybinding(msgId, gocui.KeyEsc, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		g.DeleteView(msgId)
		g.DeleteView(okId)
		g.DeleteKeybinding(okId, gocui.KeyEnter, gocui.ModNone)
		g.DeleteKeybinding(okId, gocui.KeyEsc, gocui.ModNone)
		g.DeleteKeybinding(msgId, gocui.KeyEsc, gocui.ModNone)
		g.SetCurrentView("inventory")
		return nil
	})

	return nil
}

func IsInventoryOpen() bool {
	return inventoryOpen
}
