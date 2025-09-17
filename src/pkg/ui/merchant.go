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

func updateMerchantHover(g *gocui.Gui, merchant *structures.Merchant) {
	mx, my := g.MousePosition()
	listView, _ := g.View("merchant_list")
	if listView == nil {
		return
	}
	x0, y0, x1, y1 := listView.Dimensions()
	if mx >= x0 && mx <= x1 && my >= y0 && my <= y1 {
		idx := my - y0 - 1
		if IsValidIndex(idx, len(merchant.Inventory)) && idx != merchantSelected {
			merchantSelected = idx
			g.Update(func(*gocui.Gui) error { return nil })
		}
	}
}

func attemptPurchase(g *gocui.Gui, merchant *structures.Merchant, player *structures.Player, itemIndex int) error {
	if !IsValidIndex(itemIndex, len(merchant.Inventory)) {
		return nil
	}

	entry := merchant.Inventory[itemIndex]
	item := entry.GetItem()

	if player.Money < item.Price {
		return ShowMessageWithOk(g, "merchant", "Merchant", fmt.Sprintf("Not enough gold for %s", item.Name), 50, 7)
	}
	if !player.CanAddItem(entry) {
		return ShowMessageWithOk(g, "merchant", "Merchant", "Inventory is too heavy", 50, 7)
	}

	ok := merchant.BuyItem(player, entry)
	if ok {
		if merchantSelected >= len(merchant.Inventory) && merchantSelected > 0 {
			merchantSelected = len(merchant.Inventory) - 1
		}
		return ShowMessageWithOk(g, "merchant", "Merchant", fmt.Sprintf("Purchased %s", item.Name), 50, 7)
	}
	return ShowMessageWithOk(g, "merchant", "Merchant", "Purchase failed", 50, 7)
}

func ShowMerchantMenu(merchant *structures.Merchant, player *structures.Player) {
	merchantSelected = 0
	g, _ := gocui.NewGui(gocui.OutputNormal, false)
	defer g.Close()

	g.SetManagerFunc(func(g *gocui.Gui) error { return merchantLayout(g, merchant, player) })
	merchantKeybindings(g, merchant, player)
	g.MainLoop()
}

func merchantLayout(g *gocui.Gui, merchant *structures.Merchant, player *structures.Player) error {
	maxX, maxY := g.Size()

	updateMerchantHover(g, merchant)

	if err := SetOrUpdateView(g, "merchant_title", 0, 0, maxX-1, 2, func(v *gocui.View) {
		v.Frame = false
	}, func(v *gocui.View) {
		fmt.Fprintf(v, "  Merchant • Gold: %d\n", player.Money)
	}); err != nil {
		return err
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
		if len(merchant.Inventory) == 0 {
			fmt.Fprintln(v, "(Merchant will restock soon...)")
		} else {
			lines := make([]string, 0, len(merchant.Inventory))
			for _, entry := range merchant.Inventory {
				item := entry.GetItem()
				lines = append(lines, fmt.Sprintf("%s  | Price: %d  | Rarity: %d", item.Name, item.Price, item.Rarity))
			}
			RenderListWithHighlight(v, lines, merchantSelected)
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
	BindQuitOnEsc(g)

	BindListNavigation(g, "merchant_list", &merchantSelected, func() int { return len(merchant.Inventory) })

	g.SetKeybinding("merchant_list", gocui.KeyEnter, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return attemptPurchase(g, merchant, player, merchantSelected)
	})

	EnableMouseAndSetHandler(g, func(g *gocui.Gui, v *gocui.View) error {
		mx, my := g.MousePosition()

		if listView, _ := g.View("merchant_list"); listView != nil {
			x0, y0, x1, y1 := listView.Dimensions()
			if mx >= x0 && mx <= x1 && my >= y0 && my <= y1 {
				idx := my - y0 - 1
				if IsValidIndex(idx, len(merchant.Inventory)) {
					merchantSelected = idx
					g.Update(func(*gocui.Gui) error { return nil })
					return attemptPurchase(g, merchant, player, merchantSelected)
				}
			}
		}

		buttons := []ButtonHandler{
			{"merchant_close", func(g *gocui.Gui, v *gocui.View) error { return gocui.ErrQuit }},
			{"merchant_ok", func(g *gocui.Gui, v *gocui.View) error {
				DeleteViews(g, "merchant_msg", "merchant_ok")
				return nil
			}},
		}
		return HandleMouseClickButtons(g, mx, my, buttons)
	})
	return nil
}
