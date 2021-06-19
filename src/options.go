package main

import (
	"github.com/pborman/getopt/v2"
	"log"
	"runtime"
)

type options struct {
	Bitstream string
	Build     bool
	Copy      bool
	DestDir   string
	Docker    bool
	External  bool
	ForcePull bool
	HDF       string
	NoClean   bool
	NoQSPI    bool
	Password  string
	ProjDef   string
	Release   string
	Shell     bool
	SshCopy   bool
	SshIP     string
}

func parseOptions() options {
	var opt options

	opt.DestDir = "/tmp/sd"
	opt.Password = "root"

	getopt.FlagLong(&opt.Bitstream, "bitstream", 'B', "Bitstream location")
	getopt.FlagLong(&opt.Build, "build", 'b', "Build image")
	getopt.FlagLong(&opt.Copy, "copy", 'c', "Copy image files to directory")
	getopt.FlagLong(&opt.DestDir, "dest", 'D', "Destination directory to copy the output image")
	getopt.FlagLong(&opt.Docker, "docker", 'd', "Use docker for executing the required action")
	getopt.FlagLong(&opt.External, "external", 'E', "Use external source tree")
	getopt.FlagLong(&opt.HDF, "hdf", 'H', "HDF file (will override configured bitstream)")
	getopt.FlagLong(&opt.NoClean, "no-clean", 0, "Don't remove changes on layers")
	// XXX: This should not be plain text password
	getopt.FlagLong(&opt.Password, "password", 'p', "Password for the root user")
	getopt.FlagLong(&opt.ProjDef, "project", 'P', "Project definition file")
	getopt.FlagLong(&opt.Release, "release", 'R', "Override defined release")
	getopt.FlagLong(&opt.Shell, "shell", 's', "Spawn a shell just before start compiling")
	getopt.FlagLong(&opt.SshCopy, "ssh-copy", 'S', "Copy the image remotely (by default it will copy the content to the SD and QSPI)")
	getopt.FlagLong(&opt.SshIP, "ssh-ip", 0, "Specify SSH IP where firmware will be copied")
	getopt.FlagLong(&opt.NoQSPI, "no-qspi", 0, "Do not copy the new content to QSPI flash memory")
	getopt.FlagLong(&opt.ForcePull, "force-pull", 0, "Force update of yocto and meta-layers")

	getopt.ParseV2()

	if opt.SshCopy {
		opt.Copy = true
	}

	if opt.Copy {
		if opt.HDF == "" && opt.Bitstream == "" {
			log.Print("Bitstream is required when copy flag is used")
			runtime.Goexit()
		}
	}

	return opt
}
