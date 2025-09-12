package save

import (
	"encoding/json"
	"errors"
	"os"
)

func isSaveFolderExists(saveId string) bool {
	_, err := os.Stat("saves/" + saveId)
	return err == nil
}

func CreateSaveFile(saveId, fileName string) (*os.File, error) {
	if !isSaveFolderExists(saveId) {
		err := os.Mkdir("saves/"+saveId, 0755) // 0755 is the permission for the save folder
		if err != nil {
			return nil, err
		}
	}

	path := "saves/" + saveId + "/" + fileName + ".json"
	jsonFile, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	return jsonFile, nil
}

func SaveAny(saveId, fileName string, obj interface{}) error {
	jsonFile, err := CreateSaveFile(saveId, fileName)
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	if obj == nil {
		return errors.New("object to save cannot be nil")
	}
	encoder := json.NewEncoder(jsonFile)
	encoder.SetIndent("", "  ")
	return encoder.Encode(obj)
}

func LoadAny(saveId, fileName string, obj interface{}) error {
	if !isSaveFolderExists(saveId) {
		return errors.New("save folder does not exist")
	}

	path := "saves/" + saveId + "/" + fileName + ".json"
	jsonFile, err := os.Open(path)
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	if obj == nil {
		return errors.New("object to load cannot be nil")
	}
	decoder := json.NewDecoder(jsonFile)
	return decoder.Decode(obj)
}
