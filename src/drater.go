package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

const (
	port          = ":2022"
	rootDir       = "./files"
	privateKeyPath = "id_rsa" // path to priv key for auth
)

func auth(username, password string) bool {
	expectedPassword := os.Getenv("SFTP_PASSWORD")
	if expectedPassword == "" {
		log.Fatal("SFTP_PASSWORD environment variable not set")
	}
	return password == expectedPassword
}

func handleConnection(conn net.Conn) {
	privateBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		log.Fatalf("Failed to read private key: %v", err)
	}
	privateKey, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		log.Fatalf("Failed to parse private key: %v", err)
	}

	config := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if auth(c.User(), string(pass)) {
				return nil, nil
			}
			return nil, fmt.Errorf("password rejected for %q", c.User())
		},
	}

	config.AddHostKey(privateKey)

	_, chans, reqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		log.Printf("Failed to handshake: %v", err)
		return
	}

	go ssh.DiscardRequests(reqs)

	for newChannel := range chans {
		go handleChannel(newChannel)
	}
}

func handleChannel(newChannel ssh.NewChannel) {
	channel, reqs, err := newChannel.Accept()
	if err != nil {
		log.Printf("Could not accept channel: %v", err)
		return
	}

	go ssh.DiscardRequests(reqs)

	sftpServer, err := sftp.NewServer(channel)
	if err != nil {
		log.Printf("Failed to start SFTP server: %v", err)
		return
	}

	if err := sftpServer.Serve(); err != nil {
		log.Printf("SFTP server error: %v", err)
	}
}

func main() {
	if err := os.MkdirAll(rootDir, 0755); err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}
	if os.Getenv("SFTP_PASSWORD") == "" {
		log.Fatal("SFTP_PASSWORD environment variable not set")
	}

	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to start listener: %v", err)
	}
	defer listener.Close()

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Fatalf("Failed to get IP addresses: %v", err)
	}

	fmt.Println("SFTP Server is running on the following IP addresses:")
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				fmt.Printf("IP: %s\n", ipNet.IP.String())
			}
		}
	}
	fmt.Printf("SFTP Server started on port%s\n", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		go handleConnection(conn)
	}
}
