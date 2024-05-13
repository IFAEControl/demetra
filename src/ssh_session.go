package main

import (
	"bytes"
	"context"
	scp "github.com/bramvdbogaerde/go-scp"
	"golang.org/x/crypto/ssh"
	"log"
	"os"
	"runtime"
)

type SshSession struct {
	client *ssh.Client
}

func NewSshSession(addr, password string) *SshSession {
	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", addr+":22", config)
	if err != nil {
		log.Print("Failed to dial: ", err)
		runtime.Goexit()
	}

	return &SshSession{client}
}

func (s SshSession) Run(cmd string) {
	session, err := s.client.NewSession()
	if err != nil {
		log.Print("Failed to create session: ", err)
		runtime.Goexit()
	}
	defer session.Close()

	var buff bytes.Buffer
	var bufferr bytes.Buffer
	session.Stdout = &buff
	session.Stderr = &bufferr
	if err := session.Run(cmd); err != nil {
		log.Print("Failed to run: " + bufferr.String() + " " + err.Error())
		runtime.Goexit()
	}
	//fmt.Println("Buf is:" + buff.String())
}

func (s SshSession) CopyFile(src, dst string) {
	scpClient, err := scp.NewClientBySSH(s.client)
	if err != nil {
		log.Print("Error creating new SSH session from existing connection: ", err)
		runtime.Goexit()
	}
	defer scpClient.Close()

	f, err := os.Open(src)
	LogAndExit(err)
	defer f.Close()

	err = scpClient.CopyFile(context.Background(), f, dst, "0655")
	if err != nil {
		log.Print("Error while copying file: ", err)
		runtime.Goexit()
	}
}

func (s SshSession) Close() {
	s.client.Close()
}
