package main

import (
	"log"

	"github.com/BurntSushi/toml"
)

type repo struct {
	Uri    string
	Layers []string
}

type tomlConfig struct {
	SetupDir string `toml:"setup_dir"`
	Release  string
	Machine  string
	Repo     []repo
}

func parseConfig(f string) (tomlConfig, error) {
	var config tomlConfig
	if _, err := toml.DecodeFile(f, &config); err != nil {
		log.Fatal(err)
		return tomlConfig{}, err
	}

	return config, nil
}
