package main

import (
	"bufio"
	"errors"
	"github.com/pborman/getopt/v2"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/progrium/go-basher"
)

type LocalConf struct {
}

func (c LocalConf) append(line string) {
	if c.contains(line) {
		return
	}

	f, err := os.OpenFile("build/conf/local.conf", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()
	if _, err := f.WriteString(line + "\n"); err != nil {
		log.Fatal(err)
	}
}

func (c LocalConf) contains(line string) bool {
	f, err := os.OpenFile("build/conf/local.conf", os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		if strings.Contains(scanner.Text(), line) {
			return true
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return false
}

func (c LocalConf) set(key, val string) {
	line := key + " = \"" + val + "\""

	// If value already set return without doing anytying
	if c.contains(line) {
		return
	}

	f, err := os.OpenFile("build/conf/local.conf", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()
	if _, err := f.WriteString(line + "\n"); err != nil {
		log.Fatal(err)
	}
}

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

func setupSingleLayer(b *basher.Context, uri string, layers ...string) {
	err := runCommand(b, "clone", []string{uri})
	if err != nil {
		log.Fatal(err)
	}

	for _, l := range layers {
		err = runCommand(b, "check_layer", []string{l})
		if err != nil {
			log.Fatal(err)
		}
	}
}

func setupLayers(b *basher.Context, layers []repo) {
	// first setup default layers
	default_layers := []repo{
		{"git@gitlab.pic.es:DESI-GFA/yocto/meta-dev.git", []string{"meta-dev"}},
		{"git@gitlab.pic.es:ifaecontrol/meta-ifae.git", []string{"meta-ifae"}},
		{"git://git.yoctoproject.org/meta-xilinx", []string{"meta-xilinx/meta-xilinx-bsp"}},
		{"git://git.openembedded.org/meta-openembedded",
			[]string{
				"meta-openembedded/meta-oe",
				"meta-openembedded/meta-python",
				"meta-openembedded/meta-networking",
			},
		},
	}

	for _, l := range default_layers {
		setupSingleLayer(b, l.Uri, l.Layers...)
	}

	// then setup extra layers
	for _, l := range layers {
		setupSingleLayer(b, l.Uri, l.Layers...)
	}
}

func setupYocto(b *basher.Context, cfg tomlConfig, external bool) {
	b.Export("RELEASE", cfg.Release)
	if b.HandleFuncs(os.Args) {
		os.Exit(0)
	}

	b.Source("scripts/helper_functions.sh", nil)

	err := os.MkdirAll(cfg.SetupDir, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	err = os.Chdir(cfg.SetupDir)
	if err != nil {
		log.Fatal(err)
	}

	err = runCommand(b, "clone", []string{"git://git.yoctoproject.org/poky"})
	if err != nil {
		log.Fatal(err)
	}

	err = runCommand(b, "setup_build_dir", []string{})
	if err != nil {
		log.Fatal(err)
	}

	err = runCommand(b, "rebuild_local_conf", []string{})
	if err != nil {
		log.Fatal(err)
	}

	err = runCommand(b, "checkout_machine", []string{cfg.Machine})
	if err != nil {
		log.Fatal(err)
	}

	err = os.Chdir("poky")
	if err != nil {
		log.Fatal(err)
	}

	conf := LocalConf{}
	if external {
		conf.append("INHERIT += \"externalsrc\"")

		for key, value := range cfg.Srcs {
			k := "EXTERNALSRC_pn" + key
			path := Expand(value.Path)
			conf.set(k, path)

			if value.Module {
				k := "EXTERNALSRC_BUILD_pn-" + key
				conf.set(k, path)
			}
		}
	}

	setupLayers(b, cfg.Repo)
}

func main() {
	proj_def := getopt.StringLong("project", 'P', "", "Project definition file")
	build := getopt.BoolLong("build", 'b', "", "Build image")
	release := getopt.StringLong("release", 'R', "", "Override defined release")
	external := getopt.BoolLong("external", 'E', "Use external source tree")

	getopt.Parse()

	cfg, err := parseConfig(*proj_def)
	if err != nil {
		log.Fatal(err)
	}

	if *release != "" {
		cfg.Release = *release
	}

	bash, _ := basher.NewContext("/bin/bash", false)
	bash.CopyEnv()

	setupYocto(bash, cfg, *external)

	// build
	if *build {
		err = runCommand(bash, "build", []string{})
		if err != nil {
			log.Fatal(err)
		}
	}
}
