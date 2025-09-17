package display

import (
	"fmt"
	"main/pkg/structures"
)

func handleRespawnChoice() bool {
	for {
		var input string
		fmt.Scanln(&input)

		if input == "" {
			return true
		}

		if len(input) > 0 && (input[0] == 27 || input == "q" || input == "quit" || input == "exit") {
			return false
		}

		fmt.Println("Press Enter to respawn or type 'quit' to exit...")
	}
}

func respawnPlayer(player *structures.Player) {
	player.Entity.HP = player.Entity.MaxHP / 2
	player.Entity.Alive = true

	player.Inventory = structures.Inventory{}

	player.Weapon = structures.AllWeapons["Sword"]
	player.Entity.Helmet = structures.Armors{
		Name:    "None",
		Type:    "Helmet",
		Defense: 0,
	}
	player.Entity.Chestplate = structures.Armors{
		Name:    "None",
		Type:    "Chestplate",
		Defense: 0,
	}
	player.Entity.Boots = structures.Armors{
		Name:    "None",
		Type:    "Boots",
		Defense: 0,
	}

}
