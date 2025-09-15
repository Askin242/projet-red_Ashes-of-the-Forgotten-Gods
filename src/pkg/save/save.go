package save

import (
	"encoding/json"
	"errors"
	"os"
)

var SaveId string

func isSaveFolderExists() bool {
	if SaveId == "" {
		return false
	}
	_, err := os.Stat("saves/" + SaveId)
	return err == nil
}

func CreateSaveFile(fileName string) (*os.File, error) {
	if SaveId == "" {
		return nil, errors.New("save id is not set")
	}
	if !isSaveFolderExists() {
		err := os.Mkdir("saves/"+SaveId, 0755) // 0755 is the permission for the save folder
		if err != nil {
			return nil, err
		}
	}

	path := "saves/" + SaveId + "/" + fileName + ".json"
	jsonFile, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	return jsonFile, nil
}

func SaveAny(fileName string, obj interface{}) error {
	jsonFile, err := CreateSaveFile(fileName)
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

func LoadAny(fileName string, obj interface{}) error {
	if !isSaveFolderExists() {
		return errors.New("save folder does not exist")
	}

	path := "saves/" + SaveId + "/" + fileName + ".json"
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
