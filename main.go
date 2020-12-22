package main

import (
	"github.com/pborman/getopt/v2"
	"log"
	"os"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	proj_def := getopt.StringLong("project", 'P', "", "Project definition file")
	build := getopt.BoolLong("build", 'b', "", "Build image")
	release := getopt.StringLong("release", 'R', "", "Override defined release")
	external := getopt.BoolLong("external", 'E', "Use external source tree")
	docker := getopt.BoolLong("docker", 'd', "Use docker for executing the required action")
	no_clean := getopt.BoolLong("no-clean", 0, "Don't remove changes on layers")
	shell := getopt.BoolLong("shell", 's', "Spawn a shell just before start compiling")

	// XXX: This should not be plain text password
	password := getopt.StringLong("password", 'p', "root", "Password for the root user")

	getopt.Parse()

	b := NewBash()

	if *docker {
		var args []string
		for _, v := range os.Args[1:] {
			if v != "--docker" && v != "-d" {
				args = append(args, v)
			}
		}

		// Run this program inside container and exit
		b.Source("scripts/helper_functions.sh")
		b.Run("dockerized_run", args...)
		os.Exit(0)
	}

	cfg, err := parseConfig(*proj_def)
	if err != nil {
		log.Fatal(err)
	}

	cfg.SetupDir = Expand(cfg.SetupDir)

	if *release != "" {
		cfg.Release = *release
	}

	yocto := Yocto{b, cfg, *external, *password, !*no_clean}
	yocto.setupYocto()

	// build
	if *build {
		yocto.BuildImage(*shell)
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
