package ui

import (
	"errors"
	"fmt"

	"main/pkg/structures"

	"github.com/awesome-gocui/gocui"
)

var (
	merchantSelected int
)

func ShowMerchantMenu(merchant *structures.Merchant, player *structures.Player) {
	g, _ := gocui.NewGui(gocui.OutputNormal, false)
	defer g.Close()

	g.SetManagerFunc(func(g *gocui.Gui) error { return merchantLayout(g, merchant, player) })
	merchantKeybindings(g, merchant, player)
	g.MainLoop()
}

func merchantLayout(g *gocui.Gui, merchant *structures.Merchant, player *structures.Player) error {
	maxX, maxY := g.Size()

	if v, err := g.SetView("merchant_title", 0, 0, maxX-1, 2, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Frame = false
		fmt.Fprintf(v, "  Merchant • Gold: %d\n", player.Money)
	} else {
		v.Clear()
		fmt.Fprintf(v, "  Merchant • Gold: %d\n", player.Money)
	}

	listWidth := maxX - 2
	listHeight := maxY - 6
	if listHeight < 5 {
		listHeight = 5
	}
	leftWidth := listWidth / 2
	rightStartX := 1 + leftWidth + 1
	if v, err := g.SetView("merchant_list", 1, 3, 1+leftWidth, 3+listHeight, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Title = " Items for sale "
		v.Highlight = false
	}

	if v, err := g.View("merchant_list"); err == nil {
		v.Clear()
		for i, entry := range merchant.Inventory {
			item := entry.GetItem()
			line := fmt.Sprintf("%s  | Price: %d  | Rarity: %d", item.Name, item.Price, item.Rarity)
			if i == merchantSelected {
				fmt.Fprintf(v, "\u001b[7m%s\u001b[0m\n", line)
			} else {
				fmt.Fprintln(v, line)
			}
		}
		if len(merchant.Inventory) == 0 {
			fmt.Fprintln(v, "(Merchant will restock soon...)")
		}
	}

	if v, err := g.SetView("player_inventory", rightStartX, 3, 1+listWidth, 3+listHeight, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Title = " Your inventory "
		v.Highlight = false
	}
	if v, err := g.View("player_inventory"); err == nil {
		v.Clear()
		for _, entry := range player.Inventory {
			item := entry.GetItem()
			fmt.Fprintf(v, "%s  | Rarity: %d\n", item.Name, item.Rarity)
		}
		if len(player.Inventory) == 0 {
			fmt.Fprintln(v, "(Empty)")
		}
	}

	closeBtnX := maxX - 12
	if closeBtnX < 2 {
		closeBtnX = 2
	}
	closeBtnY := maxY - 3
	if closeBtnY < 3 {
		closeBtnY = 3
	}
	createButton(g, "merchant_close", " Close ", closeBtnX, closeBtnY, 10, 2, "merchant_close")

	if len(merchant.Inventory) > 0 {
		g.SetCurrentView("merchant_list")
	}
	return nil
}

func merchantKeybindings(g *gocui.Gui, merchant *structures.Merchant, player *structures.Player) error {
	g.SetKeybinding("", gocui.KeyEsc, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return gocui.ErrQuit
	})

	g.SetKeybinding("merchant_list", gocui.KeyArrowUp, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if merchantSelected > 0 {
			merchantSelected--
		}
		return nil
	})

	g.SetKeybinding("merchant_list", gocui.KeyArrowDown, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if merchantSelected < len(merchant.Inventory)-1 {
			merchantSelected++
		}
		return nil
	})

	g.SetKeybinding("merchant_list", gocui.KeyEnter, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if len(merchant.Inventory) == 0 {
			return nil
		}
		if merchantSelected < 0 || merchantSelected >= len(merchant.Inventory) {
			return nil
		}
		entry := merchant.Inventory[merchantSelected]
		item := entry.GetItem()

		if player.Money < item.Price {
			return showMerchantMessage(g, fmt.Sprintf("Not enough gold for %s", item.Name))
		}
		if !player.CanAddItem(entry) {
			return showMerchantMessage(g, "Inventory is too heavy")
		}

		ok := merchant.BuyItem(player, entry)
		if ok {
			if merchantSelected >= len(merchant.Inventory) && merchantSelected > 0 {
				merchantSelected = len(merchant.Inventory) - 1
			}
			return showMerchantMessage(g, fmt.Sprintf("Purchased %s", item.Name))
		}
		return showMerchantMessage(g, "Purchase failed")
	})

	g.Mouse = true
	g.SetKeybinding("", gocui.MouseLeft, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		mx, my := g.MousePosition()
		if isMouseOver(g, "merchant_ok", mx, my) {
			g.DeleteView("merchant_msg")
			g.DeleteView("merchant_ok")
			return nil
		}
		if isMouseOver(g, "merchant_close", mx, my) {
			return gocui.ErrQuit
		}
		return nil
	})
	return nil
}

func showMerchantMessage(g *gocui.Gui, message string) error {
	maxX, maxY := g.Size()
	msgWidth := 50
	msgHeight := 7
	x := (maxX - msgWidth) / 2
	y := (maxY - msgHeight) / 2

	if v, err := g.SetView("merchant_msg", x, y, x+msgWidth, y+msgHeight, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Title = " Merchant "
		fmt.Fprintf(v, "\n  %s\n\n", message)
	} else {
		v.Clear()
		v.Title = " Merchant "
		fmt.Fprintf(v, "\n  %s\n\n", message)
	}

	btnX := x + msgWidth - 14
	btnY := y + msgHeight - 2
	createButton(g, "merchant_ok", " OK ", btnX, btnY-1, 10, 2, "merchant_ok")

	g.SetKeybinding("merchant_ok", gocui.KeyEnter, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		g.DeleteView("merchant_msg")
		g.DeleteView("merchant_ok")
		return nil
	})
	g.SetKeybinding("merchant_msg", gocui.KeyEsc, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		g.DeleteView("merchant_msg")
		g.DeleteView("merchant_ok")
		return nil
	})
	return nil
}
