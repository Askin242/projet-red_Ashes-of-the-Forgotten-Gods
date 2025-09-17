package save

import (
	"os"
	"path/filepath"
)

type GameConfig struct {
	Username    string `json:"username"`
	Race        string `json:"race"`
	Seed        string `json:"seed"`        // Base seed
	CurrentSeed int64  `json:"currentSeed"` // Current RNG state
}

func SaveGameConfig(config GameConfig) error {
	return SaveAny("game_config", config)
}

func LoadGameConfig() (GameConfig, error) {
	var config GameConfig
	err := LoadAny("game_config", &config)
	if err != nil {
		return GameConfig{}, err
	}

	return config, nil
}

func SetSaveID(username string) {
	SaveId = username
}

type SaveInfo struct {
	Username    string
	Race        string
	Seed        string // Base seed for display
	CurrentSeed int64  // Current seed state
}

func GetAvailableSaves() ([]SaveInfo, error) {
	var saves []SaveInfo

	if _, err := os.Stat("saves"); os.IsNotExist(err) {
		return saves, nil
	}

	entries, err := os.ReadDir("saves")
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			username := entry.Name()
			configPath := filepath.Join("saves", username, "game_config.json")

			if _, err := os.Stat(configPath); err == nil {
				oldSaveId := SaveId
				SaveId = username
				config, err := LoadGameConfig()
				SaveId = oldSaveId

				if err == nil {
					saves = append(saves, SaveInfo{
						Username:    config.Username,
						Race:        config.Race,
						Seed:        config.Seed,
						CurrentSeed: config.CurrentSeed,
					})
				}
			}
		}
	}

	return saves, nil
}

func SaveExists(username string) bool {
	configPath := filepath.Join("saves", username, "game_config.json")
	_, err := os.Stat(configPath)
	return err == nil
}
