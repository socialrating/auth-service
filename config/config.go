package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Port      string `yaml:"port"`
	MongoURI  string `yaml:"mongo_uri"`
	JWTSecret string `yaml:"jwt_secret"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
