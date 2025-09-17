package structures

import (
	"main/pkg/save"
	"time"
)

type Merchant struct {
	Entity
	Inventory       Inventory
	FirstHealBought bool
	refillTicker    *time.Ticker
	stopAutoRefill  chan struct{}
}

func (m *Merchant) AddItem(entry InventoryEntry) {
	m.Inventory = append(m.Inventory, entry)
}

func (m *Merchant) RemoveItem(entry InventoryEntry) {
	for i, item := range m.Inventory {
		if item.GetItem().Id == entry.GetItem().Id {
			m.Inventory = append(m.Inventory[:i], m.Inventory[i+1:]...)
			return
		}
	}
}

func (m *Merchant) BuyItem(player *Player, entry InventoryEntry) bool {
	if player.Money < entry.GetItem().Price {
		return false
	}

	if entry.GetItem().Price == 0 {
		m.FirstHealBought = true
	}

	player.Money -= entry.GetItem().Price
	for _, item := range m.Inventory {
		if item.GetItem().Id == entry.GetItem().Id {
			player.AddItem(item)
			m.RemoveItem(entry)
			save.SaveAny("merchant", m)
			save.SaveAny("player", player)
			return true
		}
	}
	return false
}

func (m *Merchant) Refill() {
	m.Inventory = Inventory{}

	if !m.FirstHealBought {
		m.Inventory = append(m.Inventory, GetPotion("Heal", 1, 0))
	}

	for range "123456" { // 6 items per refill + potion
		m.Inventory = append(m.Inventory, GetRandomItemByRarity())
	}
	save.SaveAny("merchant", m)
}

func InitMerchant() Merchant {
	m := Merchant{}
	err := save.LoadAny("merchant", &m)
	if err != nil {
		m = Merchant{
			Entity: Entity{
				HP:    100,
				MaxHP: 100,
				Name:  "Merchant",
			},
			Inventory:       Inventory{},
			FirstHealBought: false,
		}
		m.Refill()
		save.SaveAny("merchant", m)
	}
	m.StartAutoRefill()
	return m
}

func (m *Merchant) StartAutoRefill() {
	m.stopAutoRefill = make(chan struct{}) // Channel to stop the auto refill (required by goroutine)
	m.refillTicker = time.NewTicker(2 * time.Minute)
	go func() {
		for {
			select {
			case <-m.refillTicker.C:
				m.Refill()
			case <-m.stopAutoRefill: // Wait for channel event
				m.refillTicker.Stop()
				m.refillTicker = nil
				return
			}
		}
	}()
}

func (m *Merchant) StopAutoRefill() {
	if m.stopAutoRefill != nil {
		close(m.stopAutoRefill)
		m.stopAutoRefill = nil
	}
}
