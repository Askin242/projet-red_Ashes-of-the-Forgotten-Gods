package main

import (
	"fmt"

	structures "main/pkg/structures"
)

func main() {
	fmt.Println("=== Ashes of the Forgotten Gods - Sanity Test ===")

	// Initialize player and enemy
	player := structures.InitCharacter("John", "Human")
	enemy := structures.InitEnemy("Gruk", "Orc")

	fmt.Printf("Player %s (Race: %s) -> HP: %d/%d, Mana: %d, Level: %d\n", player.Entity.Name, player.Race.Name, player.Entity.HP, player.Entity.MaxHP, player.Mana, player.Entity.Level)
	fmt.Printf("Enemy  %s (Race: %s) -> HP: %d/%d, Level: %d\n", enemy.Entity.Name, enemy.EnemyRace.Name, enemy.Entity.HP, enemy.Entity.MaxHP, enemy.Entity.Level)

	// Level up test
	oldLevel := player.Entity.Level
	newLevel := player.LevelUp()
	fmt.Printf("Level up: %d -> %d; New HP: %d/%d, New Mana: %d\n", oldLevel, newLevel, player.Entity.HP, player.Entity.MaxHP, player.Mana)

	// Inventory capacity test
	fmt.Printf("Initial carry weight: %d/%d (items: %d)\n", player.CurrentCarryWeight(), player.MaxCarryWeight, len(player.Inventory))
	addAttempts := 10
	added := 0
	for i := 0; i < addAttempts; i++ {
		if player.AddItem(structures.AllPotions["Poison"]) {
			added++
		} else {
			break
		}
	}
	fmt.Printf("Tried adding %d Poison potions; actually added: %d\n", addAttempts, added)
	fmt.Printf("Post-add carry weight: %d/%d (items: %d)\n", player.CurrentCarryWeight(), player.MaxCarryWeight, len(player.Inventory))

	// Armor set bonus and damage mitigation test
	player.Entity.Helmet = structures.AllHelmets["StormBringer"]
	player.Entity.Chestplate = structures.AllChestplates["StormBringer"]
	player.Entity.Boots = structures.AllBoots["StormBringer"]
	setBonus := structures.GetSetBonusDefense(player.Entity)
	totalDefense := player.Entity.Helmet.Defense + player.Entity.Chestplate.Defense + player.Entity.Boots.Defense + setBonus
	fmt.Printf("Equipped full StormBringer set. Set bonus: %d, Total defense: %d\n", setBonus, totalDefense)

	// Direct damage test on player entity to ensure HP decreases correctly
	beforeHP := player.Entity.HP
	player.Entity.TakeDamage(50)
	afterHP := player.Entity.HP
	fmt.Printf("Player took 50 raw damage -> HP: %d -> %d (Alive: %v)\n", beforeHP, afterHP, player.Entity.Alive)

	// Direct damage test on enemy entity
	enemyBeforeHP := enemy.Entity.HP
	enemy.Entity.TakeDamage(30)
	enemyAfterHP := enemy.Entity.HP
	fmt.Printf("Enemy took 30 raw damage -> HP: %d -> %d (Alive: %v)\n", enemyBeforeHP, enemyAfterHP, enemy.Entity.Alive)

	// Spell reference check
	fireball := structures.AllSpells["Fireball"]
	fmt.Printf("Spell loaded: %s (Damage: %d, Cost: %d)\n", fireball.Name, fireball.Damage, fireball.Cost)

	fmt.Println("=== Sanity Test Complete ===")
}
