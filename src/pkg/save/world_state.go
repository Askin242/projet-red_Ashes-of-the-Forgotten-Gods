package save

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type WorldState struct {
	CurrentLevel int        `json:"currentLevel"`
	PlayerX      int        `json:"playerX"`
	PlayerY      int        `json:"playerY"`
	Mobs         []Position `json:"mobs"`
}
