package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Server    Server `yaml:"server"`
	MongoURI  string `yaml:"mongo_uri"`
	JWTSecret string `yaml:"jwt_secret"`
}

type Server struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
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
