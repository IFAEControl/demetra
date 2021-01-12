package main

import (
	"fmt"
	"log"
	"os"
)

type Yocto struct {
	b          *Bash
	cfg        tomlConfig
	external   bool
	password   string
	clean      bool
	demetraDir string
}

func (y Yocto) setupSingleLayer(uri, release string, layers ...string) {
	y.setupRepo(uri, "", release)
	for _, l := range layers {
		y.b.Run("check_layer", l)
	}
}

func (y Yocto) setupLayers(layers []repo, release string) {
	// first setup default layers
	default_layers := []repo{
		{"git@gitlab.pic.es:ifaecontrol/meta-dev.git", []string{"meta-dev"}},
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
		y.setupSingleLayer(l.Uri, release, l.Layers...)
	}

	// then setup extra layers
	for _, l := range layers {
		y.setupSingleLayer(l.Uri, release, l.Layers...)
	}
}

func (y Yocto) setupBuildDir(sd string) {

	build_dir := fmt.Sprint(sd, "/build")

	cmd := fmt.Sprint("cd ", sd, "; source ./oe-init-build-env > /dev/null")
	y.b.Run("bash", "-c", cmd)

	if !Exists(build_dir) {
		log.Fatal("Error when creating poky build directory")
	}
}

func (y Yocto) rebuildLocalCfg(sd string) {
	local_conf := fmt.Sprint(sd, "/build/conf/local.conf")
	// ignore error, if config not exist will be created
	os.Remove(local_conf)
	y.setupBuildDir(sd)
}

func (y Yocto) cloneRepo(repo, directory string) string {
	if directory != "" {
		if Exists(directory) {
			return directory
		}

		y.b.Run("git", "clone", repo, directory)
	} else {
		directory = GetStem(repo)
		y.cloneRepo(repo, directory)
	}

	return directory
}

func (y Yocto) setupRepo(repo, directory, release string) {
	if directory == "" {
		directory = GetStem(repo)
	}

	if !Exists(directory) {
		directory = y.cloneRepo(repo, directory)
	}

	old_dir, _ := os.Getwd()
	err := os.Chdir(directory)
	if err != nil {
		log.Fatal(err)
	}

	if y.clean {
		y.b.Run("git", "checkout", "--", ".")
		y.b.Run("git", "clean", "-fd")
	}

	y.b.Run("git", "pull")
	y.b.Run("checkout_repository", release)

	err = os.Chdir(old_dir)
	if err != nil {
		log.Fatal(err)
	}
}

func (y Yocto) setupYocto() {
	y.b.Source("scripts/helper_functions.sh")

	y.setupRepo("git://git.yoctoproject.org/poky", y.cfg.SetupDir, y.cfg.Release)
	y.rebuildLocalCfg(y.cfg.SetupDir)

	err := os.Chdir(y.cfg.SetupDir)
	if err != nil {
		log.Fatal(err)
	}

	y.b.Run("checkout_machine", y.cfg.Machine)
	y.b.Run("set_password", y.password)

	conf := NewLocalConf()
	defer conf.Close()

	if y.external {
		conf.append("INHERIT += \"externalsrc\"")

		for key, value := range y.cfg.Srcs {
			k := "EXTERNALSRC_pn" + key
			path := Expand(value.Path)
			conf.set(k, path)

			if value.Module {
				k := "EXTERNALSRC_BUILD_pn-" + key
				conf.set(k, path)
			}
		}
	}

	// This is only used in gatesgarth branch but it doesn't hurt
	conf.set("HDF_BASE", "file://")
	conf.set("HDF_PATH", y.demetraDir + "/resources/latest.hdf")

	y.setupLayers(y.cfg.Repo, y.cfg.Release)
}

func (y Yocto) BuildImage(shell bool) {
	if shell {
		y.b.Run("bash")
	}
	y.b.Run("build")
}
