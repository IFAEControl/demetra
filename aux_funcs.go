package main

import (
	"log"
	"os"
	"os/user"
	"path"
	"strings"
)

func Expand(path string) string {
	usr, err := user.Current()
	if err != nil {
		log.Fatalln("Can not get current user: ", err)
	}

	return strings.Replace(path, "~", usr.HomeDir, 1)
}

func Exists(name string) bool {
	if _, err := os.Stat(name); os.IsNotExist(err) {
		return false
	}
	return true
}

func GetStem(uri string) string {
	fname := path.Base(uri)
	return strings.Split(fname, ".")[0]
}
