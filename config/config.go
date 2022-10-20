package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

func GetConfig() (Config, error) {
	configuration := Config{}
	var path string
	if runtime.GOOS == "windows" {
		ex, err := os.Executable()
		if err != nil {
			panic(err)
		}
		exPath := filepath.Dir(ex)

		path = filepath.Join(exPath, "config", "config.json")
	} else {
		path = "config/config.json"
	}

	log.Println(path)
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
