package config

import (
	"encoding/json"
	"fmt"
	"os"
)

func GetConfig() (Config, error) {
	configuration := Config{}
	path := "config/config.json"
	_, err := os.Stat(path)
	if err == nil {
		file, _ := os.Open(path)

		defer file.Close()

		decoder := json.NewDecoder(file)
		errorDecode := decoder.Decode(&configuration)
		if errorDecode != nil {
			fmt.Println("Error: ", errorDecode)
		}
		return configuration, errorDecode
	}
	return configuration, err
}
