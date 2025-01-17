package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const configPath = "./config/config.yml"

type Config struct {
	Port string `yaml:"port"`
}

func ParseConfig() Config {
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
