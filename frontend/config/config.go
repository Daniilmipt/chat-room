package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const (
	CONFIG_ENV        = "CONFIG_PATH"
	configPathDefault = "./config/config.yml"
)

type Config struct {
	Frontend FrontConfig   `yaml:"frontend"`
	Backend  BackendConfig `yaml:"backend"`
}

type FrontConfig struct {
	Port string `yaml:"port"`
}

type BackendConfig struct {
	Host   string `yaml:"host"`
	Port   string `yaml:"port"`
	PeerID string `yaml:"peer-id"`
}

func ParseConfig() Config {
	configPath := os.Getenv(CONFIG_ENV)
	if configPath == "" {
		configPath = configPathDefault
	}

	filename, _ := filepath.Abs(configPath)
	fmt.Println(filename)
	fmt.Println(configPath)
	yamlFile, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	var cfg Config
	if err = yaml.Unmarshal(yamlFile, &cfg); err != nil {
		panic(err)
	}
	return cfg
}
