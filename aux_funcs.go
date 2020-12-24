package main

import (
	"fmt"
	"io"
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

func Copy(src, dst string) (err error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}
