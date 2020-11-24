package main

import (
	//"flag"
	"log"
	"os"

	"github.com/progrium/go-basher"
)

func main() {
	cfg, err := parseConfig()

	bash, _ := basher.NewContext("/bin/bash", false)
	bash.Export("RELEASE", cfg.Release)
	if bash.HandleFuncs(os.Args) {
		os.Exit(0)
	}

	bash.Source("scripts/helper_functions.sh", nil)
	status, err := bash.Run("clone", []string{"git://git.yoctoproject.org/poky"})
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(status)
}
