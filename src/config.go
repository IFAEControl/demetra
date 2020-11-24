package main

import (
	"log"

	"github.com/BurntSushi/toml"
)

type tomlConfig struct {
	Release string
}

func parseConfig() (tomlConfig, error) {
	var config tomlConfig
	if _, err := toml.DecodeFile("example.toml", &config); err != nil {
		log.Fatal(err)
		return tomlConfig{}, err
	}

	return config, nil
}
