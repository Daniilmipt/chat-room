package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const (
	CONFIG_ENV        = "CONFIG_PATH"
	configPathDefault = "./config/config.yml"
)

type Config struct {
	General General
	Api     ApiConfig `yaml:"chat-api"`
}

type General struct {
	Port string `yaml:"port"`
}

type ApiConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

func ParseConfig() Config {
	configPath := os.Getenv(CONFIG_ENV)
	if configPath == "" {
		configPath = configPathDefault
	}

	filename, _ := filepath.Abs(configPath)
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
