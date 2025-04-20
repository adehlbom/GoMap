package main

import (
	"fmt"

	"golang.org/x/crypto/ssh"
)

func test_ssh(ip_address string, port int) {
	fmt.Println("Testing to connect to SSH server...")
	var hostKey ssh.PublicKey

	config := &ssh.ClientConfig{
		User: "",
		Auth: []ssh.AuthMethod{
			ssh.Password(""),
		},
		HostKeyCallback: ssh.FixedHostKey(hostKey),
	}
	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", ip_address, port), config)

	if err != nil {
		fmt.Println(err)
	}
	conn.Close()

}


func test_html(ip_address string, port int) {
}
