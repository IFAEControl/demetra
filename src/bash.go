package main

import (
	"bytes"
	"errors"
	"github.com/progrium/go-basher"
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

func (b Bash) runCommandWithOutput(cmd string, args []string) (string, error) {
	var buff bytes.Buffer
	b.ctx.Stdout = &buff
	status, err := b.ctx.Run(cmd, args)
	if err != nil {
		return "", err
	}

	if status != 0 {
		return "", errors.New("Unknown return number: " + strconv.Itoa(status))
	}

	b.ctx.Stdout = os.Stdout

	return buff.String(), nil
}

func (b Bash) Source(script string) {
	b.ctx.Source(script, nil)
}

func (b Bash) Export(key, value string) {
	b.ctx.Export(key, value)
}

func (b Bash) Run(cmd string, args ...string) {
	err := b.runCommand(cmd, args)
	LogAndExit(err)
}

func (b Bash) RunWithOutput(cmd string, args ...string) string {
	stdout, err := b.runCommandWithOutput(cmd, args)
	LogAndExit(err)

	return stdout
}

func NewBash() *Bash {
	ctx, _ := basher.NewContext("/bin/bash", false)
	ctx.CopyEnv()

	if ctx.HandleFuncs(os.Args) {
		os.Exit(0)
	}

	return &Bash{ctx}
}
