package main

import (
	"errors"
	"github.com/pborman/getopt/v2"
	"log"
	"os"
	"strconv"

	"github.com/progrium/go-basher"
)

func runCommand(b *basher.Context, cmd string, args []string) error {
	status, err := b.Run(cmd, args)
	if err != nil {
		return err
	}

	if status != 0 {
		return errors.New("Unknown return number: " + strconv.Itoa(status))
	}

	return nil
}

func setupSingleLayer(b *basher.Context, layer_uri, layer_name string) {
	err := runCommand(b, "clone", []string{layer_uri})
	if err != nil {
		log.Fatal(err)
	}

	err = runCommand(b, "check_layer", []string{layer_name})
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	proj_def := getopt.StringLong("project", 'P', "", "Project definition file")
	getopt.Parse()

	cfg, err := parseConfig(*proj_def)
	if err != nil {
		log.Fatal(err)
	}

	bash, _ := basher.NewContext("/bin/bash", true)
	bash.CopyEnv()
	bash.Export("RELEASE", cfg.Release)
	if bash.HandleFuncs(os.Args) {
		os.Exit(0)
	}

	bash.Source("scripts/helper_functions.sh", nil)
	err = runCommand(bash, "clone", []string{"git://git.yoctoproject.org/poky"})
	if err != nil {
		log.Fatal(err)
	}

	err = runCommand(bash, "setup_build_dir", []string{})
	if err != nil {
		log.Fatal(err)
	}

	err = runCommand(bash, "checkout_machine", []string{cfg.Machine})
	if err != nil {
		log.Fatal(err)
	}

	err = os.Chdir("poky")
	if err != nil {
		log.Fatal(err)
	}

	// default meta layers
	default_layers := []layer{
		{"meta-dev", "git@gitlab.pic.es:DESI-GFA/yocto/meta-dev.git"},
		{"meta-ifae", "git@gitlab.pic.es:ifaecontrol/meta-ifae.git"},
		{"meta-openembedded/meta-oe", "git://git.openembedded.org/meta-openembedded"},
		{"meta-openembedded/meta-python", "git://git.openembedded.org/meta-openembedded"},
		{"meta-openembedded/meta-networking", "git://git.openembedded.org/meta-openembedded"},
		{"meta-xilinx/meta-xilinx-bsp", "git://git.yoctoproject.org/meta-xilinx"},
	}

	for _, l := range default_layers {
		setupSingleLayer(bash, l.Uri, l.Name)
	}

	// extra meta layers
	for _, l := range cfg.Layer {
		setupSingleLayer(bash, l.Uri, l.Name)
	}

	// build
	err = runCommand(bash, "build", []string{})
	if err != nil {
		log.Fatal(err)
	}
}
