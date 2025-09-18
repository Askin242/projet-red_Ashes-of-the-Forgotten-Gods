package display

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"main/pkg/fight"
	"main/pkg/gmgmap"
	"main/pkg/save"
	"main/pkg/structures"
	"main/pkg/ui"

	"github.com/awesome-gocui/gocui"
)

func spawnEntities(m *gmgmap.Map, rng *rand.Rand) {
	entities := m.Layer("Entities")
	ground := m.Layer("Ground")

	playerExists := false
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			if entities.GetTile(x, y) == gmgmap.Player {
				playerExists = true
				break
			}
		}
		if playerExists {
			break
		}
	}

	var validSpawns [][]int
	for y := 1; y < m.Height-1 && y <= 32; y++ { // Don't spawn below y=32
		for x := 1; x < m.Width-2; x++ { // Leave room for 2-character wide entities
			tile1 := ground.GetTile(x, y)
			tile2 := ground.GetTile(x+1, y)
			entityTile1 := entities.GetTile(x, y)
			entityTile2 := entities.GetTile(x+1, y)

			// Both tiles must be valid floor/room tiles and empty
			if (tile1 == gmgmap.Room || tile1 == gmgmap.Room2 || tile1 == gmgmap.Floor) &&
				(tile2 == gmgmap.Room || tile2 == gmgmap.Room2 || tile2 == gmgmap.Floor) &&
				entityTile1 == gmgmap.Nothing && entityTile2 == gmgmap.Nothing {
				validSpawns = append(validSpawns, []int{x, y})
			}
		}
	}

	if len(validSpawns) == 0 {
		fmt.Println("Warning: No valid spawn locations found!")
		return
	}

	for i := range validSpawns {
		j := rng.Intn(i + 1)
		validSpawns[i], validSpawns[j] = validSpawns[j], validSpawns[i]
	}

	spawnIndex := 0

	if !playerExists && spawnIndex < len(validSpawns) {
		spawn := validSpawns[spawnIndex]
		entities.SetTile(spawn[0], spawn[1], gmgmap.Player)
		entities.SetTile(spawn[0]+1, spawn[1], gmgmap.Player) // Player is 2 chars wide
		spawnIndex++
		fmt.Printf("Player at: (%d, %d) - (%d, %d)\n", spawn[0], spawn[1], spawn[0]+1, spawn[1])
	}

	if spawnIndex < len(validSpawns) {
		spawn := validSpawns[spawnIndex]
		entities.SetTile(spawn[0], spawn[1], gmgmap.Merchant)
		entities.SetTile(spawn[0]+1, spawn[1], gmgmap.Merchant) // Merchant is 2 chars wide
		spawnIndex++
		fmt.Printf("Merchant at: (%d, %d) - (%d, %d)\n", spawn[0], spawn[1], spawn[0]+1, spawn[1])
	}

	if spawnIndex < len(validSpawns) {
		spawn := validSpawns[spawnIndex]
		entities.SetTile(spawn[0], spawn[1], gmgmap.Blacksmith)
		entities.SetTile(spawn[0]+1, spawn[1], gmgmap.Blacksmith) // Blacksmith is 2 chars wide
		spawnIndex++
		fmt.Printf("Blacksmith at: (%d, %d) - (%d, %d)\n", spawn[0], spawn[1], spawn[0]+1, spawn[1])
	}

	numMobs := rng.Intn(8) + 8
	for i := 0; i < numMobs && spawnIndex < len(validSpawns); i++ {
		spawn := validSpawns[spawnIndex]
		entities.SetTile(spawn[0], spawn[1], gmgmap.Mob)
		entities.SetTile(spawn[0]+1, spawn[1], gmgmap.Mob) // Mobs are 2 chars wide
		spawnIndex++
		fmt.Printf("Mob %d at: (%d, %d) - (%d, %d)\n", i+1, spawn[0], spawn[1], spawn[0]+1, spawn[1])
	}

	fmt.Printf("Total entities spawned: %d\n", spawnIndex)
}

func findPlayer(m *gmgmap.Map) (int, int) {
	entities := m.Layer("Entities")
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			if entities.GetTile(x, y) == gmgmap.Player {
				return x, y
			}
		}
	}
	return -1, -1
}

func canMoveTo(m *gmgmap.Map, x, y int) bool {
	// Check bounds for both tiles (player is 2 characters wide)
	if x < 0 || x+1 >= m.Width || y < 0 || y >= m.Height {
		return false
	}

	if y > 32 { // else the player can go outside the view cuz of bottom
		return false
	}

	ground := m.Layer("Ground")
	structures := m.Layer("Structures")
	entities := m.Layer("Entities")

	for i := 0; i < 2; i++ {
		groundTile := ground.GetTile(x+i, y)
		structureTile := structures.GetTile(x+i, y)
		entityTile := entities.GetTile(x+i, y)

		validGround := (groundTile == gmgmap.Room || groundTile == gmgmap.Room2 ||
			groundTile == gmgmap.Floor)

		invalidStructure := (structureTile == gmgmap.Wall || structureTile == gmgmap.Wall2)

		blockedByEntity := entityTile != gmgmap.Nothing &&
			entityTile != gmgmap.Mob &&
			entityTile != gmgmap.Merchant &&
			entityTile != gmgmap.Blacksmith &&
			entityTile != gmgmap.Player

		if !validGround || invalidStructure || blockedByEntity {
			return false
		}
	}

	return true
}

func movePlayer(m *gmgmap.Map, oldX, oldY, newX, newY int) {
	entities := m.Layer("Entities")

	entities.SetTile(oldX, oldY, gmgmap.Nothing)
	entities.SetTile(oldX+1, oldY, gmgmap.Nothing)

	entities.SetTile(newX, newY, gmgmap.Player)
	entities.SetTile(newX+1, newY, gmgmap.Player)
}

type GameState struct {
	gameMap          *gmgmap.Map
	playerX, playerY int
	currentLevel     int
	maps             map[int]*gmgmap.Map
	gui              *gocui.Gui
	player           *structures.Player
}

var gameState *GameState

func gameLayout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	if v, err := g.SetView("game", 0, 0, maxX-1, maxY-4, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Frame = false
		v.Wrap = false
		updateGameView(v)

		if _, err := g.SetCurrentView("game"); err != nil {
			return err
		}
	}

	if v, err := g.SetView("status", 0, maxY-4, maxX-1, maxY-1, 0); err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return err
		}
		v.Frame = true
		v.Title = " Status "
		updateStatusView(v)
	}

	return nil
}

func updateGameView(v *gocui.View) {
	if gameState == nil || gameState.gameMap == nil {
		return
	}

	v.Clear()

	m := gameState.gameMap
	ground := m.Layer("Ground")
	entities := m.Layer("Entities")

	for y := 0; y < m.Height; y++ {
		line := ""
		skipNext := false

		for x := 0; x < m.Width; x++ {
			if skipNext {
				skipNext = false
				continue
			}

			entityTile := entities.GetTile(x, y)
			if entityTile != gmgmap.Nothing && gmgmap.IsDoubleWidthEntity(entityTile) {
				groundTile := ground.GetTile(x, y)
				line += gmgmap.GetEntitySymbolWithBackground(entityTile, groundTile)
				skipNext = true
				continue
			}

			rendered := false
			for i := len(m.Layers) - 1; i >= 0; i-- {
				l := m.Layers[i]
				tile := l.GetTile(x, y)
				if i == 0 || tile != gmgmap.Nothing {
					if l.Name == "Entities" && tile != gmgmap.Nothing {
						groundTile := ground.GetTile(x, y)
						line += gmgmap.GetEntitySymbolWithBackground(tile, groundTile)
					} else {
						line += gmgmap.GetTileSymbol(tile)
					}
					rendered = true
					break
				}
			}
			if !rendered {
				line += " "
			}
		}
		fmt.Fprintln(v, line)
	}
}

func updateStatusView(v *gocui.View) {
	if gameState == nil {
		return
	}

	v.Clear()

	var difficultyDesc string
	switch {
	case gameState.currentLevel == 0:
		difficultyDesc = "Easy"
	case gameState.currentLevel <= 2:
		difficultyDesc = "Moderate"
	case gameState.currentLevel <= 5:
		difficultyDesc = "Hard"
	case gameState.currentLevel <= 8:
		difficultyDesc = "Very Hard"
	default:
		difficultyDesc = "Extreme"
	}

	xpProgress := gameState.player.XP % 100
	barLength := 20
	filledLength := int(float64(xpProgress) / 100.0 * float64(barLength))

	xpBar := "["
	for i := 0; i < barLength; i++ {
		if i < filledLength {
			xpBar += "="
		} else {
			xpBar += "-"
		}
	}
	xpBar += "]"

	fmt.Fprintf(v, "HP: %d/%d | Gold: %d | Mana: %d | Level: %d | XP: %s %d/100 | Dungeon: %d (%s)",
		gameState.player.Entity.HP, gameState.player.Entity.MaxHP, gameState.player.Money,
		gameState.player.Mana, gameState.player.Entity.Level, xpBar, xpProgress, gameState.currentLevel, difficultyDesc)
	fmt.Fprint(v, "\nZ=Up S=Down Q=Left D=Right F=Stairs E=Inventory X=Exit ESC=Menu | 😊=You 😈=Enemies 👑=Merchant ⚒️=Blacksmith")
}

func moveUp(g *gocui.Gui, v *gocui.View) error {
	return tryMove(g, 0, -1)
}

func moveDown(g *gocui.Gui, v *gocui.View) error {
	return tryMove(g, 0, 1)
}

func moveLeft(g *gocui.Gui, v *gocui.View) error {
	return tryMove(g, -1, 0)
}

func moveRight(g *gocui.Gui, v *gocui.View) error {
	return tryMove(g, 1, 0)
}

func exitGame(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func openInventory(g *gocui.Gui, v *gocui.View) error {
	if gameState == nil || gameState.player == nil {
		return nil
	}
	return ui.ShowInventory(g, gameState.player)
}

func generateMapForLevel(level int, rng *rand.Rand) *gmgmap.Map {
	width, height := 150, 40
	splits := 3
	minRoomSize := 15
	corridorWidth := 3

	m := gmgmap.NewBSPInterior(rng, func(_ *gmgmap.Map) {}, width, height, splits, minRoomSize, corridorWidth)

	structures := m.Layer("Structures")

	if level == 0 {
		for y := 0; y < m.Height; y++ {
			for x := 0; x < m.Width; x++ {
				if structures.GetTile(x, y) == gmgmap.StairsUp {
					structures.SetTile(x, y, gmgmap.StairsDown)
				}
			}
		}
	}

	return m
}

func useStairs(g *gocui.Gui, v *gocui.View) error {
	if gameState == nil {
		return nil
	}

	structuresX := gameState.gameMap.Layer("Structures")
	leftTile := structuresX.GetTile(gameState.playerX, gameState.playerY)
	rightTile := leftTile
	if gameState.playerX+1 < gameState.gameMap.Width {
		rightTile = structuresX.GetTile(gameState.playerX+1, gameState.playerY)
	}

	var currentTile rune
	if gmgmap.IsStairs(leftTile) {
		currentTile = leftTile
	} else if gmgmap.IsStairs(rightTile) {
		currentTile = rightTile
	} else {
		return nil
	}

	var newLevel int
	if currentTile == gmgmap.StairsUp {
		newLevel = gameState.currentLevel + 1
		if newLevel > 0 {
			return nil
		}
	} else if currentTile == gmgmap.StairsDown {
		newLevel = gameState.currentLevel - 1
	} else {
		return nil
	}

	oldMap := gameState.gameMap
	if oldMap != nil {
		oldMap.Layer("Entities").SetTile(gameState.playerX, gameState.playerY, gmgmap.Nothing)
	}

	if gameState.maps[newLevel] == nil {
		rng := structures.GetRNG()
		gameState.maps[newLevel] = generateMapForLevel(newLevel, rng)
	}

	gameState.currentLevel = newLevel
	gameState.gameMap = gameState.maps[newLevel]

	entities := gameState.gameMap.Layer("Entities")
	ground := gameState.gameMap.Layer("Ground")

	var targetStairs rune = gmgmap.StairsUp
	if currentTile == gmgmap.StairsUp {
		targetStairs = gmgmap.StairsDown
	}

	structuresX = gameState.gameMap.Layer("Structures")
	found := false

	for y := 1; y < gameState.gameMap.Height-1 && !found; y++ {
		for x := 1; x < gameState.gameMap.Width-1 && !found; x++ {
			if structuresX.GetTile(x, y) == targetStairs {
				for dy := -1; dy <= 1 && !found; dy++ {
					for dx := -1; dx <= 1 && !found; dx++ {
						newX, newY := x+dx, y+dy
						if newX >= 0 && newX+1 < gameState.gameMap.Width && newY >= 0 && newY < gameState.gameMap.Height {
							groundTile1 := ground.GetTile(newX, newY)
							groundTile2 := ground.GetTile(newX+1, newY)
							entityTile1 := entities.GetTile(newX, newY)
							entityTile2 := entities.GetTile(newX+1, newY)
							if (groundTile1 == gmgmap.Room || groundTile1 == gmgmap.Room2 || groundTile1 == gmgmap.Floor) &&
								(groundTile2 == gmgmap.Room || groundTile2 == gmgmap.Room2 || groundTile2 == gmgmap.Floor) &&
								entityTile1 == gmgmap.Nothing && entityTile2 == gmgmap.Nothing {
								gameState.playerX = newX
								gameState.playerY = newY
								entities.SetTile(newX, newY, gmgmap.Player)
								entities.SetTile(newX+1, newY, gmgmap.Player) // Player is 2 chars wide
								found = true
							}
						}
					}
				}
			}
		}
	}

	if !found {
		for y := 1; y < gameState.gameMap.Height-1 && y <= 32 && !found; y++ { // Don't spawn below y=32
			for x := 1; x < gameState.gameMap.Width-2 && !found; x++ { // Leave room for 2-character wide player
				groundTile1 := ground.GetTile(x, y)
				groundTile2 := ground.GetTile(x+1, y)
				entityTile1 := entities.GetTile(x, y)
				entityTile2 := entities.GetTile(x+1, y)
				if (groundTile1 == gmgmap.Room || groundTile1 == gmgmap.Room2 || groundTile1 == gmgmap.Floor) &&
					(groundTile2 == gmgmap.Room || groundTile2 == gmgmap.Room2 || groundTile2 == gmgmap.Floor) &&
					entityTile1 == gmgmap.Nothing && entityTile2 == gmgmap.Nothing {
					gameState.playerX = x
					gameState.playerY = y
					entities.SetTile(x, y, gmgmap.Player)
					entities.SetTile(x+1, y, gmgmap.Player) // Player is 2 chars wide
					found = true
				}
			}
		}
	}

	hasEntities := false
	for y := 0; y < gameState.gameMap.Height && !hasEntities; y++ {
		for x := 0; x < gameState.gameMap.Width && !hasEntities; x++ {
			tile := entities.GetTile(x, y)
			if tile != gmgmap.Nothing && tile != gmgmap.Player {
				hasEntities = true
			}
		}
	}

	if !hasEntities {
		rng := structures.GetRNG()
		spawnEntities(gameState.gameMap, rng)
	}

	_ = save.SaveWorldState(save.WorldState{CurrentLevel: gameState.currentLevel, PlayerX: gameState.playerX, PlayerY: gameState.playerY})

	g.Update(func(g *gocui.Gui) error {
		gameView, _ := g.View("game")
		statusView, _ := g.View("status")
		if gameView != nil {
			updateGameView(gameView)
		}
		if statusView != nil {
			updateStatusView(statusView)
		}
		return nil
	})

	return nil
}

func createRandomEnemy() *structures.Enemy {
	rng := structures.GetRNG()

	dungeonLevel := gameState.currentLevel

	// Different enemy types appear at different depths
	var enemyRaces []string
	var enemyNames map[string][]string

	switch {
	case dungeonLevel >= 8:
		// Deep levels: All enemy types, but more dangerous ones are more common
		enemyRaces = []string{"Skeleton", "Skeleton", "Orc", "Orc", "Goblin"}
		enemyNames = map[string][]string{
			"Orc":      {"Gruk the Destroyer", "Thok Bonecrusher", "Morg the Terrible", "Ugluk Deathbringer"},
			"Skeleton": {"Ancient Bones", "Death's Rattles", "Cursed Marrow", "Lich Skull"},
			"Goblin":   {"Shadow Sneaky", "Poison Stabby", "Blood Greedy", "Vile Nasty"},
		}
	case dungeonLevel >= 5:
		// Mid-deep levels: Stronger variations
		enemyRaces = []string{"Orc", "Skeleton", "Goblin", "Orc"}
		enemyNames = map[string][]string{
			"Orc":      {"Gruk the Fierce", "Thok Ironjaw", "Morg the Brutal", "Ugluk Warbringer"},
			"Skeleton": {"Cursed Bones", "Wailing Rattles", "Dark Marrow", "Hollow Skull"},
			"Goblin":   {"Cunning Sneaky", "Deadly Stabby", "Vicious Greedy", "Cruel Nasty"},
		}
	case dungeonLevel >= 2:
		// Mid levels: Standard enemies with some variety
		enemyRaces = []string{"Orc", "Skeleton", "Goblin"}
		enemyNames = map[string][]string{
			"Orc":      {"Gruk the Bold", "Thok Strongarm", "Morg the Wild", "Ugluk Raider"},
			"Skeleton": {"Restless Bones", "Clattering Rattles", "Dry Marrow", "Grinning Skull"},
			"Goblin":   {"Sly Sneaky", "Quick Stabby", "Hungry Greedy", "Mean Nasty"},
		}
	default:
		// Surface and shallow levels: Weaker, more basic enemies
		enemyRaces = []string{"Goblin", "Goblin", "Orc", "Skeleton"}
		enemyNames = map[string][]string{
			"Orc":      {"Gruk", "Thok", "Morg", "Ugluk"},
			"Skeleton": {"Bones", "Rattles", "Marrow", "Skull"},
			"Goblin":   {"Sneaky", "Stabby", "Greedy", "Nasty"},
		}
	}

	race := enemyRaces[rng.Intn(len(enemyRaces))]
	names := enemyNames[race]
	name := names[rng.Intn(len(names))]

	// Use the new scaled enemy initialization based on current dungeon level
	enemy := structures.InitScaledEnemy(name, race, dungeonLevel)

	// Add level-based prefix to enemy name to indicate difficulty
	if dungeonLevel > 0 {
		var prefix string
		switch {
		case dungeonLevel >= 10:
			prefix = "Legendary "
		case dungeonLevel >= 7:
			prefix = "Elite "
		case dungeonLevel >= 5:
			prefix = "Veteran "
		case dungeonLevel >= 3:
			prefix = "Seasoned "
		case dungeonLevel >= 1:
			prefix = "Experienced "
		}
		enemy.Entity.Name = prefix + enemy.Entity.Name
	}

	return &enemy
}

var merchant = structures.InitMerchant()
var blacksmith = structures.InitCraftingBlacksmith()

func tryMove(g *gocui.Gui, dx, dy int) error {
	if gameState == nil {
		return nil
	}

	newX := gameState.playerX + dx
	newY := gameState.playerY + dy

	if canMoveTo(gameState.gameMap, newX, newY) {
		entities := gameState.gameMap.Layer("Entities")

		entityTile1 := entities.GetTile(newX, newY)
		entityTile2 := gmgmap.Nothing
		if newX+1 < gameState.gameMap.Width {
			entityTile2 = entities.GetTile(newX+1, newY)
		}

		if entityTile1 == gmgmap.Mob || entityTile2 == gmgmap.Mob {
			enemy := createRandomEnemy()

			for cx := newX - 1; cx <= newX+2; cx++ {
				if cx >= 0 && cx < gameState.gameMap.Width {
					if entities.GetTile(cx, newY) == gmgmap.Mob {
						entities.SetTile(cx, newY, gmgmap.Nothing)
					}
				}
			}

			g.Close()
			ui.ClearScreen()

			playerWon := fight.StartFight(gameState.player, enemy)

			if !gameState.player.Entity.Alive {
				ui.ClearScreen()
				fmt.Println("=== GAME OVER ===")
				fmt.Printf("Your character %s has fallen in battle!\n", gameState.player.Entity.Name)
				fmt.Println()
				fmt.Println("You can respawn with 50% health, but you'll lose all your equipment.")
				fmt.Println("Press Enter to respawn or Esc to quit...")

				if handleRespawnChoice() {
					respawnPlayer(gameState.player)
					save.SaveAny("player", gameState.player)

					ui.ClearScreen()
					fmt.Println("You have been revived! You wake up at the entrance with only the basics...")
					fmt.Println("Press any key to continue...")
					fmt.Scanln()

					entities.SetTile(newX, newY, gmgmap.Nothing)
					if newX+1 < gameState.gameMap.Width {
						entities.SetTile(newX+1, newY, gmgmap.Nothing)
					}

					if playerWon {
						for cx := newX - 1; cx <= newX+2; cx++ {
							if cx >= 0 && cx < gameState.gameMap.Width {
								if entities.GetTile(cx, newY) == gmgmap.Mob {
									entities.SetTile(cx, newY, gmgmap.Nothing)
								}
							}
						}
					}
					movePlayer(gameState.gameMap, gameState.playerX, gameState.playerY, newX, newY)
					gameState.playerX = newX
					gameState.playerY = newY

					ui.ClearScreen()
					return restartGameLoop()
				} else {
					return nil // Quit the game
				}
			}

			movePlayer(gameState.gameMap, gameState.playerX, gameState.playerY, newX, newY)
			gameState.playerX = newX
			gameState.playerY = newY
			_ = save.SaveWorldState(save.WorldState{CurrentLevel: gameState.currentLevel, PlayerX: gameState.playerX, PlayerY: gameState.playerY})

			ui.ClearScreen()
			return restartGameLoop()
		}

		if entityTile1 == gmgmap.Merchant || entityTile2 == gmgmap.Merchant {
			g.Close()
			ui.ClearScreen()

			ui.ShowMerchantMenu(merchant, gameState.player)

			save.SaveAny("player", gameState.player)

			ui.ClearScreen()
			return restartGameLoop()
		}

		if entityTile1 == gmgmap.Blacksmith || entityTile2 == gmgmap.Blacksmith {
			g.Close()
			ui.ClearScreen()

			ui.ShowBlacksmithMenu(blacksmith, gameState.player)

			save.SaveAny("player", gameState.player)

			ui.ClearScreen()
			return restartGameLoop()
		}

		movePlayer(gameState.gameMap, gameState.playerX, gameState.playerY, newX, newY)
		gameState.playerX = newX
		gameState.playerY = newY
		_ = save.SaveWorldState(save.WorldState{CurrentLevel: gameState.currentLevel, PlayerX: gameState.playerX, PlayerY: gameState.playerY})

		structuresX := gameState.gameMap.Layer("Structures")
		leftTile := structuresX.GetTile(gameState.playerX, gameState.playerY)
		rightTile := leftTile
		if gameState.playerX+1 < gameState.gameMap.Width {
			rightTile = structuresX.GetTile(gameState.playerX+1, gameState.playerY)
		}
		if gmgmap.IsStairs(leftTile) || gmgmap.IsStairs(rightTile) {
			return useStairs(g, nil)
		}

		g.Update(func(g *gocui.Gui) error {
			gameView, _ := g.View("game")
			statusView, _ := g.View("status")
			if gameView != nil {
				updateGameView(gameView)
			}
			if statusView != nil {
				updateStatusView(statusView)
			}
			return nil
		})
	}

	return nil
}

func setupKeybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("", 'z', gocui.ModNone, moveUp); err != nil {
		return err
	}
	if err := g.SetKeybinding("", 'Z', gocui.ModNone, moveUp); err != nil {
		return err
	}
	if err := g.SetKeybinding("", 's', gocui.ModNone, moveDown); err != nil {
		return err
	}
	if err := g.SetKeybinding("", 'S', gocui.ModNone, moveDown); err != nil {
		return err
	}
	if err := g.SetKeybinding("", 'q', gocui.ModNone, moveLeft); err != nil {
		return err
	}
	if err := g.SetKeybinding("", 'Q', gocui.ModNone, moveLeft); err != nil {
		return err
	}
	if err := g.SetKeybinding("", 'd', gocui.ModNone, moveRight); err != nil {
		return err
	}
	if err := g.SetKeybinding("", 'D', gocui.ModNone, moveRight); err != nil {
		return err
	}

	if err := g.SetKeybinding("", 'e', gocui.ModNone, openInventory); err != nil {
		return err
	}
	if err := g.SetKeybinding("", 'E', gocui.ModNone, openInventory); err != nil {
		return err
	}
	if err := g.SetKeybinding("", 'x', gocui.ModNone, exitGame); err != nil {
		return err
	}
	if err := g.SetKeybinding("", 'X', gocui.ModNone, exitGame); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, exitGame); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyEsc, gocui.ModNone, exitGame); err != nil {
		return err
	}

	return nil
}

func restartGameLoop() error {
	if gameState == nil {
		fmt.Println("Error: gameState is nil!")
		return fmt.Errorf("gameState is nil")
	}

	g, err := gocui.NewGui(gocui.OutputNormal, true)
	if err != nil {
		fmt.Printf("Error creating GUI: %v\n", err)
		return err
	}

	gameState.gui = g

	g.Cursor = false
	g.SetManagerFunc(gameLayout)

	if err := setupKeybindings(g); err != nil {
		fmt.Printf("Error setting up keybindings: %v\n", err)
		g.Close()
		return err
	}

	if err := g.MainLoop(); err != nil && !errors.Is(err, gocui.ErrQuit) {
		fmt.Printf("Error in main loop: %v\n", err)
		return err
	}

	g.Close()
	return nil
}

func ShowNewGameForm(player *structures.Player) {
	ui.ClearScreen()
	fmt.Println("It's your first time playing this!")
	fmt.Println("Would you like to do a training fight? You won't lose any HP or anything; it's just for testing.")
	fmt.Print("Start training fight now? (y/n): ")
	var ans string
	fmt.Scanln(&ans)
	ans = strings.TrimSpace(strings.ToLower(ans))
	if len(ans) > 0 && ans[0] == 'y' {
		ui.ClearScreen()
		trainingPlayer := *player
		enemy := structures.InitScaledEnemy("Training Dummy", "Goblin", 0)
		fight.StartFight(&trainingPlayer, &enemy)
		fmt.Println("\nTraining finished. Press Enter to continue...")
		fmt.Scanln()
	}
}

func StartGame(username, race, seedStr string) {
	save.SetSaveID(username)

	// Initialize the seed system
	if seedStr != "" {
		structures.InitializeSeed(seedStr)
	} else {
		// Generate a random seed if none provided
		randomSeed := fmt.Sprintf("%d", time.Now().UTC().UnixNano())
		structures.InitializeSeed(randomSeed)
		seedStr = randomSeed
	}

	fmt.Printf("Starting game for %s (%s) with seed %s\n", username, race, seedStr)
	rng := structures.GetRNG()

	m := generateMapForLevel(0, rng)
	spawnEntities(m, rng)

	player := structures.InitCharacter(username, race)

	if player.IsFirstLogin {
		ShowNewGameForm(&player)
		player.IsFirstLogin = false
		save.SaveAny("player", player)
	}

	fmt.Println("Starting game... Use ZQSD to move, F to use stairs, ESC for menu")
	fmt.Println("Walk over merchants/blacksmiths to interact with them")

	for {
		err := startGameLoopWithPlayer(m, &player)
		if errors.Is(err, ErrReturnToMainMenu) {
			fmt.Println("Returning to main menu...")
			ui.ShowMainMenu()
		} else {
			return // Exit completely
		}
	}
}

func startGameLoopWithPlayer(m *gmgmap.Map, player *structures.Player) error {
	playerX, playerY := findPlayer(m)
	if playerX == -1 || playerY == -1 {
		fmt.Println("Error: Player not found on map!")
		return fmt.Errorf("player not found on map")
	}

	maps := make(map[int]*gmgmap.Map)
	maps[0] = m

	gameState = &GameState{
		gameMap:      m,
		playerX:      playerX,
		playerY:      playerY,
		currentLevel: 0,
		maps:         maps,
		player:       player,
	}

	if ws, err := save.LoadWorldState(); err == nil {
		if ws.CurrentLevel <= 0 {
			if gameState.maps[ws.CurrentLevel] == nil {
				rng := structures.GetRNG()
				gameState.maps[ws.CurrentLevel] = generateMapForLevel(ws.CurrentLevel, rng)
			}
			gameState.currentLevel = ws.CurrentLevel
			gameState.gameMap = gameState.maps[ws.CurrentLevel]
			entities := gameState.gameMap.Layer("Entities")
			for y := 0; y < gameState.gameMap.Height; y++ {
				for x := 0; x < gameState.gameMap.Width; x++ {
					if entities.GetTile(x, y) == gmgmap.Player {
						entities.SetTile(x, y, gmgmap.Nothing)
					}
				}
			}
			gameState.playerX = ws.PlayerX
			gameState.playerY = ws.PlayerY
			if gameState.playerX+1 < gameState.gameMap.Width {
				entities.SetTile(gameState.playerX, gameState.playerY, gmgmap.Player)
				entities.SetTile(gameState.playerX+1, gameState.playerY, gmgmap.Player)
			}

			hasEntities := false
			for y := 0; y < gameState.gameMap.Height && !hasEntities; y++ {
				for x := 0; x < gameState.gameMap.Width && !hasEntities; x++ {
					tile := entities.GetTile(x, y)
					if tile != gmgmap.Nothing && tile != gmgmap.Player {
						hasEntities = true
					}
				}
			}
			if !hasEntities {
				rng := structures.GetRNG()
				spawnEntities(gameState.gameMap, rng)
			}
		}
	}

	g, err := gocui.NewGui(gocui.OutputNormal, true)
	if err != nil {
		fmt.Printf("Error creating GUI: %v\n", err)
		return err
	}
	defer g.Close()

	gameState.gui = g

	g.Cursor = false
	g.SetManagerFunc(gameLayout)

	if err := setupKeybindingsWithPlayer(g); err != nil {
		fmt.Printf("Error setting up keybindings: %v\n", err)
		return err
	}
	if err := setupGameMenuKeybindings(g); err != nil {
		fmt.Printf("Error setting up menu keybindings: %v\n", err)
		return err
	}

	err = g.MainLoop()
	if err != nil && !errors.Is(err, gocui.ErrQuit) && !errors.Is(err, ErrReturnToMainMenu) {
		fmt.Printf("Error in main loop: %v\n", err)
	}

	return err
}

func setupKeybindingsWithPlayer(g *gocui.Gui) error {
	if err := g.SetKeybinding("", 'z', gocui.ModNone, moveUp); err != nil {
		return err
	}
	if err := g.SetKeybinding("", 'Z', gocui.ModNone, moveUp); err != nil {
		return err
	}
	if err := g.SetKeybinding("", 's', gocui.ModNone, moveDown); err != nil {
		return err
	}
	if err := g.SetKeybinding("", 'S', gocui.ModNone, moveDown); err != nil {
		return err
	}
	if err := g.SetKeybinding("", 'q', gocui.ModNone, moveLeft); err != nil {
		return err
	}
	if err := g.SetKeybinding("", 'Q', gocui.ModNone, moveLeft); err != nil {
		return err
	}
	if err := g.SetKeybinding("", 'd', gocui.ModNone, moveRight); err != nil {
		return err
	}
	if err := g.SetKeybinding("", 'D', gocui.ModNone, moveRight); err != nil {
		return err
	}

	if err := g.SetKeybinding("", 'e', gocui.ModNone, openInventory); err != nil {
		return err
	}
	if err := g.SetKeybinding("", 'E', gocui.ModNone, openInventory); err != nil {
		return err
	}
	if err := g.SetKeybinding("", 'x', gocui.ModNone, exitGame); err != nil {
		return err
	}
	if err := g.SetKeybinding("", 'X', gocui.ModNone, exitGame); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, exitGame); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyEsc, gocui.ModNone, exitGame); err != nil {
		return err
	}

	return nil
}
