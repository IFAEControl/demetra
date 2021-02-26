package main

import (
	"bufio"
	"log"
	"os"
	"strings"
)

type LocalConf struct {
	file *os.File
}

func NewLocalConf() *LocalConf {
	f, err := os.OpenFile("build/conf/local.conf", os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err)
	}

	return &LocalConf{f}
}

func (c LocalConf) Close() {
	c.file.Close()
}

func (c LocalConf) add(line string) {
	// If value already set return without doing anytying
	if c.contains(line) {
		return
	}

	if _, err := c.file.WriteString(line + "\n"); err != nil {
		log.Fatal(err)
	}
}

func (c LocalConf) contains(line string) bool {
	scanner := bufio.NewScanner(c.file)

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

func (c LocalConf) append(key, val string) {
	line := key + "_append = \" " + val + "\""
	c.add(line)
}

func (c LocalConf) set(key, val string) {
	line := key + " = \"" + val + "\""
	c.add(line)
}

func (c LocalConf) setDefault(key, val string) {
	line := key + " ?= \"" + val + "\""
	c.add(line)
}
