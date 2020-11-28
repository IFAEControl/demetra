package main

import (
	"log"
	"os/user"
	"strings"
)

func Expand(path string) string {
	usr, err := user.Current()
	if err != nil {
		log.Fatalln("Can not get current user: ", err)
	}

	return strings.Replace(path, "~", usr.HomeDir, 1)
}
