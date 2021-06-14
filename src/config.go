package main

import (
	"log"

	"github.com/BurntSushi/toml"
)

type src struct {
	Module bool
	Path   string
}

type repo struct {
	Uri    string
	Layers []string
}

type tomlConfig struct {
	SetupDir string `toml:"setup_dir"`
	Release  string
	Machine  string
	Packages []string
	Repo     []repo
	Srcs     map[string]src
}

func parseConfig(f string) (tomlConfig, error) {
	var config tomlConfig
	if _, err := toml.DecodeFile(f, &config); err != nil {
		log.Print(err)
		return tomlConfig{}, err
	}

	return config, nil
}
