package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Regarding the bit vs bin files (we currently convert from bit format to bin)
	// bit and bin are identical at bit level the only difference is that the one have
	// a header and the second one doesn't. It supposed that the FFFFF after the header
	// resets the FPGA, so any possible change introduced by reading the header is ignored.
	// If that is true, I don't know why we needed to convert it.

	opt := parseOptions()

	b := NewBash()

	demetraDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	cfg, err := parseConfig(opt.ProjDef)
	if err != nil {
		log.Fatal(err)
	}

	if opt.Docker {
		var args []string
		for _, v := range os.Args[1:] {
			if v != "--docker" && v != "-d" {
				args = append(args, v)
			}
		}

		var volumes string
		for _, v := range cfg.Srcs {
			p := Expand(v.Path)
			volumes += " -v " + p + ":" + p
		}

		d := GetSstateCacheDir()
		volumes += " -v " + d + ":" + d

		// Run this program inside container and exit
		b.Source("scripts/helper_functions.sh")
		b.Export("DOCKER_MOUNT_ARGS", volumes)
		b.Run("dockerized_run", args...)
		os.Exit(0)
	}

	cfg.SetupDir, err = filepath.Abs(Expand(cfg.SetupDir))
	if err != nil {
		log.Fatal(err)
	}

	if opt.Release != "" {
		cfg.Release = opt.Release
	}

	if opt.HDF != "" {
		if cfg.Release == "gatesgarth" {
			err = Copy(opt.HDF, "resources/latest.hdf")
			if err != nil {
				panic(err)
			}
		} else {
			// Prepare directory where firmware image will be hold temporarily
			dir, err := ioutil.TempDir("", "demetra")
			if err != nil {
				log.Fatal(err)
			}
			defer os.RemoveAll(dir)

			paths, err := Unzip(opt.HDF, dir)
			if err != nil {
				log.Fatal(err)
			}

			for _, p := range paths {
				if filepath.Ext(p) == ".bit" {
					opt.Bitstream = p
					break
				}
			}
			log.Println("Using bitstream: " + opt.Bitstream)
		}
	}

	yocto := Yocto{b, cfg, opt.External, opt.Password, !opt.NoClean, opt.ForcePull, demetraDir}
	yocto.setupYocto()

	// build
	if opt.Build {
		yocto.BuildImage(opt.Shell)
	}

	// TODO: Implement copy script in Go
	if opt.Copy {
		//Copy(*bitstream, dest_dir+"/fpga.bit")
		b.Run("../scripts/copy.sh", opt.DestDir, "build/tmp/deploy/images/", "", cfg.Machine, opt.Bitstream, "false")
	}

	if opt.SshCopy {
		b.Run("../scripts/ssh-copy.sh", "build/tmp/deploy/images/", cfg.Machine, opt.Bitstream, opt.Password, opt.SshIP, strconv.FormatBool(opt.NoQSPI))
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
