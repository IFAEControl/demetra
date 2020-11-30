package main

import (
	"errors"
	"github.com/progrium/go-basher"
	"log"
	"os"
	"strconv"
)

type Bash struct {
	ctx *basher.Context
}

func (b Bash) runCommand(cmd string, args []string) error {
	status, err := b.ctx.Run(cmd, args)
	if err != nil {
		return err
	}

	if status != 0 {
		return errors.New("Unknown return number: " + strconv.Itoa(status))
	}

	return nil
}

func (b Bash) Source(script string) {
	b.ctx.Source(script, nil)
}

func (b Bash) Export(key, value string) {
	b.ctx.Export(key, value)
}

func (b Bash) Run(cmd string, args ...string) {
	err := b.runCommand(cmd, args)
	if err != nil {
		log.Fatal(err)
	}
}

func NewBash() *Bash {
	ctx, _ := basher.NewContext("/bin/bash", false)
	ctx.CopyEnv()

	if ctx.HandleFuncs(os.Args) {
		os.Exit(0)
	}

	return &Bash{ctx}
}
