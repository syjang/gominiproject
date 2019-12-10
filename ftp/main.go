package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func main() {
	user := "sftpuser"
	pass := "test1234"

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(pass),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		// HostKeyCallback: ssh.FixedHostKey(hostKey),
	}

	client, err := ssh.Dial("tcp", "localhost:22", config)
	if err != nil {
		panic("Failed to dial: " + err.Error())
	}
	fmt.Println("Successfully connected to ssh server.")

	// open an SFTP session over an existing ssh connection.
	sftp, err := sftp.NewClient(client)
	if err != nil {
		log.Fatal(err)
	}
	defer sftp.Close()

	srcPath := "/sftpuser/test/"
	dstPath := "./"
	filename := "t3.jpg"

	srcFile, err := os.Open(dstPath + filename)
	if err != nil {
		log.Fatal(err)
	}
	defer srcFile.Close()

	// Open the source file
	dstFile, err := sftp.Create(srcPath + filename)
	if err != nil {
		log.Fatal(err)
	}
	defer srcFile.Close()

	// Copy the file
	io.Copy(dstFile, srcFile)
}
