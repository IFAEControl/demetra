package main

import (
	"bytes"
	scp "github.com/bramvdbogaerde/go-scp"
	"golang.org/x/crypto/ssh"
	"log"
	"os"
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
		log.Fatal("Failed to dial: ", err)
	}

	return &SshSession{client}
}

func (s SshSession) Run(cmd string) {
	session, err := s.client.NewSession()
	if err != nil {
		log.Fatal("Failed to create session: ", err)
	}
	defer session.Close()

	var buff bytes.Buffer
	var bufferr bytes.Buffer
	session.Stdout = &buff
	session.Stderr = &bufferr
	if err := session.Run(cmd); err != nil {
		log.Fatal("Failed to run: " + bufferr.String() + " " + err.Error())
	}
	//fmt.Println("Buf is:" + buff.String())
}

func (s SshSession) CopyFile(src, dst string) {
	scpClient, err := scp.NewClientBySSH(s.client)
	if err != nil {
		log.Fatal("Error creating new SSH session from existing connection: ", err)
	}
	defer scpClient.Close()

	f, err := os.Open(src)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	err = scpClient.CopyFile(f, dst, "0655")
	if err != nil {
		log.Fatal("Error while copying file: ", err)
	}
}

func (s SshSession) Close() {
	s.client.Close()
}
