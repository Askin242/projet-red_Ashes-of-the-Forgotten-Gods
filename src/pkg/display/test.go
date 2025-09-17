package display

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
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
	if x < 0 || x >= m.Width || y < 0 || y >= m.Height {
		return false
	}

	if y > 32 { // else the player can go outsite the view cuz of bottom
		return false
	}

	ground := m.Layer("Ground")
	structures := m.Layer("Structures")

	groundTile := ground.GetTile(x, y)
	structureTile := structures.GetTile(x, y)

	validGround := (groundTile == gmgmap.Room || groundTile == gmgmap.Room2 ||
		groundTile == gmgmap.Floor)

	invalidStructure := (structureTile == gmgmap.Wall || structureTile == gmgmap.Wall2)

	return validGround && !invalidStructure
}

func movePlayer(m *gmgmap.Map, oldX, oldY, newX, newY int) {
	entities := m.Layer("Entities")

	entities.SetTile(oldX, oldY, gmgmap.Nothing)

	entities.SetTile(newX, newY, gmgmap.Player)
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

func clearConsole() {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls")
	default:
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

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
	fmt.Fprintf(v, "HP: %d/%d | Gold: %d | Mana: %d | Position: (%d, %d) | Level: %d",
		gameState.player.Entity.HP, gameState.player.Entity.MaxHP, gameState.player.Money,
		gameState.player.Mana, gameState.playerX, gameState.playerY, gameState.currentLevel)
	fmt.Fprint(v, "\nZ=Up S=Down Q=Left D=Right F=Use Stairs ESC=Menu | 😊=You 😈=Enemies 👑=Merchant ⚒️=Blacksmith")
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

	structures := gameState.gameMap.Layer("Structures")
	currentTile := structures.GetTile(gameState.playerX, gameState.playerY)

	if !gmgmap.IsStairs(currentTile) {
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
		seed := time.Now().UTC().UnixNano()
		rng := rand.New(rand.NewSource(seed))
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

	structures = gameState.gameMap.Layer("Structures")
	found := false

	for y := 1; y < gameState.gameMap.Height-1 && !found; y++ {
		for x := 1; x < gameState.gameMap.Width-1 && !found; x++ {
			if structures.GetTile(x, y) == targetStairs {
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
		seed := time.Now().UTC().UnixNano()
		rng := rand.New(rand.NewSource(seed))
		spawnEntities(gameState.gameMap, rng)
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

	return nil
}

func createRandomEnemy() *structures.Enemy {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	enemyRaces := []string{"Orc", "Skeleton", "Goblin"}
	enemyNames := map[string][]string{
		"Orc":      {"Gruk", "Thok", "Morg", "Ugluk"},
		"Skeleton": {"Bones", "Rattles", "Marrow", "Skull"},
		"Goblin":   {"Sneaky", "Stabby", "Greedy", "Nasty"},
	}

	race := enemyRaces[rng.Intn(len(enemyRaces))]
	names := enemyNames[race]
	name := names[rng.Intn(len(names))]

	enemy := structures.InitEnemy(name, race)

	enemy.Entity.Level = -gameState.currentLevel + 1

	weapons := []string{"Sword", "Axe", "DoubleAxes", "Spear"}
	enemy.Weapon = structures.AllWeapons[weapons[rng.Intn(len(weapons))]]

	return &enemy
}

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

			if entityTile1 == gmgmap.Mob {
				entities.SetTile(newX, newY, gmgmap.Nothing)
			}
			if entityTile2 == gmgmap.Mob && newX+1 < gameState.gameMap.Width {
				entities.SetTile(newX+1, newY, gmgmap.Nothing)
			}

			g.Close()
			clearConsole()

			fight.StartFight(gameState.player, enemy)

			if !gameState.player.Entity.Alive {
				clearConsole()
				fmt.Println("=== GAME OVER ===")
				fmt.Printf("Your character %s has fallen in battle!\n", gameState.player.Entity.Name)
				fmt.Println()
				fmt.Println("You can respawn with 50% health, but you'll lose all your equipment and inventory.")
				fmt.Println("Press Enter to respawn or Esc to quit...")

				if handleRespawnChoice() {
					respawnPlayer(gameState.player)
					save.SaveAny("player", gameState.player)

					clearConsole()
					fmt.Println("You have been revived! You wake up at the entrance with only the basics...")
					fmt.Println("Press any key to continue...")
					fmt.Scanln()

					movePlayer(gameState.gameMap, gameState.playerX, gameState.playerY, newX, newY)
					gameState.playerX = newX
					gameState.playerY = newY

					clearConsole()
					return restartGameLoop()
				} else {
					return nil // Quit the game
				}
			}

			movePlayer(gameState.gameMap, gameState.playerX, gameState.playerY, newX, newY)
			gameState.playerX = newX
			gameState.playerY = newY

			clearConsole()
			return restartGameLoop()
		}

		if entityTile1 == gmgmap.Merchant || entityTile2 == gmgmap.Merchant {
			g.Close()
			clearConsole()

			merchant := structures.InitMerchant()

			ui.ShowMerchantMenu(&merchant, gameState.player)

			save.SaveAny("player", gameState.player)

			clearConsole()
			return restartGameLoop()
		}

		if entityTile1 == gmgmap.Blacksmith || entityTile2 == gmgmap.Blacksmith {
			g.Close()
			clearConsole()

			blacksmith := structures.InitCraftingBlacksmith()
			err := save.LoadAny("blacksmith_job", &blacksmith.Current)
			if err != nil {
				blacksmith.Current = nil
			}

			ui.ShowBlacksmithMenu(&blacksmith, gameState.player)

			save.SaveAny("player", gameState.player)

			clearConsole()
			return restartGameLoop()
		}

		movePlayer(gameState.gameMap, gameState.playerX, gameState.playerY, newX, newY)
		gameState.playerX = newX
		gameState.playerY = newY

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

	if err := g.SetKeybinding("", 'f', gocui.ModNone, useStairs); err != nil {
		return err
	}
	if err := g.SetKeybinding("", 'F', gocui.ModNone, useStairs); err != nil {
		return err
	}

	if err := g.SetKeybinding("", 'e', gocui.ModNone, exitGame); err != nil {
		return err
	}
	if err := g.SetKeybinding("", 'E', gocui.ModNone, exitGame); err != nil {
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

func startGameLoop(m *gmgmap.Map) {
	playerX, playerY := findPlayer(m)
	if playerX == -1 || playerY == -1 {
		fmt.Println("Error: Player not found on map!")
		return
	}

	maps := make(map[int]*gmgmap.Map)
	maps[0] = m

	player := &structures.Player{
		Entity: structures.Entity{
			Name:       "Hero",
			HP:         100,
			MaxHP:      100,
			Alive:      true,
			Level:      1,
			Helmet:     structures.GetRandomArmorByType("Helmet"),
			Chestplate: structures.GetRandomArmorByType("Chestplate"),
			Boots:      structures.GetRandomArmorByType("Boots"),
		},
		Race:           structures.Human,
		Weapon:         structures.AllWeapons["Sword"],
		Mana:           50,
		Money:          100,
		MaxCarryWeight: 100,
	}

	gameState = &GameState{
		gameMap:      m,
		playerX:      playerX,
		playerY:      playerY,
		currentLevel: 0,
		maps:         maps,
		player:       player,
	}

	g, err := gocui.NewGui(gocui.OutputNormal, true)
	if err != nil {
		fmt.Printf("Error creating GUI: %v\n", err)
		return
	}
	defer g.Close()

	gameState.gui = g

	g.Cursor = false
	g.SetManagerFunc(gameLayout)

	if err := setupKeybindings(g); err != nil {
		fmt.Printf("Error setting up keybindings: %v\n", err)
		return
	}

	if err := g.MainLoop(); err != nil && !errors.Is(err, gocui.ErrQuit) {
		fmt.Printf("Error in main loop: %v\n", err)
	}
}

func StartGame(username, race, seedStr string) {
	save.SetSaveID(username)

	var seedVal int64
	if seedStr != "" {
		for _, char := range seedStr {
			seedVal = seedVal*31 + int64(char)
		}
	} else {
		seedVal = time.Now().UTC().UnixNano()
	}

	fmt.Printf("Starting game for %s (%s) with seed %s\n", username, race, seedStr)
	rng := rand.New(rand.NewSource(seedVal))

	m := generateMapForLevel(0, rng)
	spawnEntities(m, rng)

	player := structures.InitCharacter(username, race)

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

func Display() {
	seed := time.Now().UTC().UnixNano()
	fmt.Println("Using seed", seed)
	rng := rand.New(rand.NewSource(seed))

	m := generateMapForLevel(0, rng)

	spawnEntities(m, rng)

	fmt.Println("Starting game... Use ZQSD to move, F to use stairs, E to quit")
	startGameLoop(m)
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

	if err := g.SetKeybinding("", 'f', gocui.ModNone, useStairs); err != nil {
		return err
	}
	if err := g.SetKeybinding("", 'F', gocui.ModNone, useStairs); err != nil {
		return err
	}

	if err := g.SetKeybinding("", 'e', gocui.ModNone, exitGame); err != nil {
		return err
	}
	if err := g.SetKeybinding("", 'E', gocui.ModNone, exitGame); err != nil {
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
