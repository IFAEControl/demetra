package main

import (
	"github.com/pborman/getopt/v2"
	"log"
	"os"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Regarding the bit vs bin files (we currently convert from bit format to bin)
	// bit and bin are identical at bit level the only difference is that the one have
	// a header and the second one doesn't. It supposed that the FFFFF after the header
	// resets the FPGA, so any possible change introduced by reading the header is ignored.
	// If that is true, I don't know why we needed to convert it.

	bitstream := getopt.StringLong("bitstream", 'B', "", "Bitstream location")
	build := getopt.BoolLong("build", 'b', "", "Build image")
	copy := getopt.BoolLong("copy", 'c', "Copy image files to directory")
	dest_dir_arg := getopt.StringLong("dest", 'D', "/tmp/sd", "Destination directory to copy the output image")
	docker := getopt.BoolLong("docker", 'd', "Use docker for executing the required action")
	external := getopt.BoolLong("external", 'E', "Use external source tree")
	no_clean := getopt.BoolLong("no-clean", 0, "Don't remove changes on layers")
	// XXX: This should not be plain text password
	password := getopt.StringLong("password", 'p', "root", "Password for the root user")
	proj_def := getopt.StringLong("project", 'P', "", "Project definition file")
	release := getopt.StringLong("release", 'R', "", "Override defined release")
	shell := getopt.BoolLong("shell", 's', "Spawn a shell just before start compiling")

	// create options go

	getopt.Parse()

	dest_dir := *dest_dir_arg

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

	/*if dest_dir == "" {
		// Prepare directory where firmware image will be hold temporarily
		dest_dir, err = ioutil.TempDir("", "demetra")
		if err != nil {
			log.Fatal(err)
		}
		defer os.RemoveAll(dest_dir)
	}*/

	// TODO: Implement copy script in Go
	if *bitstream != "" {
		Copy(*bitstream, dest_dir+"/fpga.bit")
	}

	if *copy {
		b.Run("../scripts/copy.sh", dest_dir, "build/tmp/deploy/images/", "", cfg.Machine, *bitstream, "false")
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
