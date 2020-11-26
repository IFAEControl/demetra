package main

import (
	"log"

	"github.com/BurntSushi/toml"
)

type layer struct {
	Name string
	Uri  string
}

type tomlConfig struct {
	Release string
	Machine string
	Layer   []layer
}

func parseConfig(f string) (tomlConfig, error) {
	var config tomlConfig
	if _, err := toml.DecodeFile(f, &config); err != nil {
		log.Fatal(err)
		return tomlConfig{}, err
	}

	return config, nil
}
