package main

import (
	"github.com/pborman/getopt/v2"
	"log"
)

type options struct {
	Bitstream string
	Build     bool
	Copy      bool
	DestDir   string
	Docker    bool
	External  bool
	HDF       string
	NoClean   bool
	Password  string
	ProjDef   string
	Release   string
	Shell     bool
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

	getopt.ParseV2()

	if opt.Copy {
		if opt.HDF == "" && opt.Bitstream == "" {
			log.Fatal("Bitstream is required when copy flag is used")
		}
	}

	return opt
}
