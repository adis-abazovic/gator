package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	CurrentUserName string `json:"current_user_name"`
	DbUrl           string `json:"db_url"`
}

const configFileName = ".gatorconfig.json"

func Read() (Config, error) {

	path, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("error reading file '%s'", path)
	}

	cfg := Config{}
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return Config{}, fmt.Errorf("error unmarshaling data")
	}

	return cfg, nil
}

func (cfg *Config) SetUser(userName string) {

	cfg.CurrentUserName = userName
	write(cfg)
}

func getConfigFilePath() (string, error) {

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	path := homeDir + "/" + configFileName
	return path, nil
}

func write(cfg *Config) error {
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	path, err := getConfigFilePath()
	if err != nil {
		return err
	}

	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return fmt.Errorf("error: writing config file failed")
	}

	return nil
}
