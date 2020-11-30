package main

import (
	"github.com/pborman/getopt/v2"
	"log"
	"os"
	"fmt"
)

func setupSingleLayer(b *Bash, uri string, layers ...string) {
	b.Run("clone", uri)
	for _, l := range layers {
		b.Run("check_layer", l)
	}
}

func setupLayers(b *Bash, layers []repo) {
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
		setupSingleLayer(b, l.Uri, l.Layers...)
	}

	// then setup extra layers
	for _, l := range layers {
		setupSingleLayer(b, l.Uri, l.Layers...)
	}
}

func setupBuildDir(b *Bash, sd string) {
	build_dir := fmt.Sprint(sd, "/build")

	if !Exists(build_dir) {
		cmd := fmt.Sprint("cd ", sd, "; source ./oe-init-build-env > /dev/null")
		b.Run("bash", "-c", cmd)
	}

	if !Exists(build_dir) {
		log.Fatal("Error when creating poky build directory")
	}
}

func rebuildLocalCfg(b *Bash, sd string) {
	local_conf := fmt.Sprint(sd, "/build/conf/local.conf")
	// ignore error, if config not exist will be created
	os.Remove(local_conf)
	setupBuildDir(b, sd)
}

func setupYocto(b *Bash, cfg tomlConfig, external bool, password string) {
	b.Export("RELEASE", cfg.Release)
	b.Source("scripts/helper_functions.sh")

	b.Run("clone", "git://git.yoctoproject.org/poky", cfg.SetupDir)
	rebuildLocalCfg(b, cfg.SetupDir)

	err := os.Chdir(cfg.SetupDir)
	if err != nil {
		log.Fatal(err)
	}

	b.Run("checkout_machine", cfg.Machine)
	b.Run("set_password", password)

	conf := NewLocalConf()
	defer conf.Close()

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
	old_dir, _ := os.Getwd()

	proj_def := getopt.StringLong("project", 'P', "", "Project definition file")
	build := getopt.BoolLong("build", 'b', "", "Build image")
	release := getopt.StringLong("release", 'R', "", "Override defined release")
	external := getopt.BoolLong("external", 'E', "Use external source tree")
	docker := getopt.BoolLong("docker", 'd', "Use docker for executing the required action")

	// XXX: This should not be plain text password
	password := getopt.StringLong("password", 'p', "root", "Password for the root user")

	getopt.Parse()

	cfg, err := parseConfig(*proj_def)
	if err != nil {
		log.Fatal(err)
	}

	if *release != "" {
		cfg.Release = *release
	}

	b := NewBash()

	setupYocto(b, cfg, *external, *password)

	// build
	if *build {
		if *docker {
			err = os.Chdir(old_dir)
			if err != nil {
				log.Fatal(err)
			}

			b.Run("build_docker", "-P "+*proj_def, "-b")
		} else {
			b.Run("build")
		}
	}
}

/*package main

import (
	"context"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

func main() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "test1234",
		Cmd:   []string{"echo", "hello world"},
		Tty:   false,
	}, nil, nil, nil, "")
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		panic(err)
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
}

*/
