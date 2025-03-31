// fix this entire mess some day

package main

import (
    "fmt"
    "log"
    "net"
    "os"
    "path/filepath"

    "github.com/pkg/sftp"
    "golang.org/x/crypto/ssh"
)

const (
    port          = ":2022"
    rootDir       = "./files"
    privateKeyPath = "id_rsa"
)

func auth(username, password string) bool {
    expectedPassword := os.Getenv("SFTP_PASSWORD")
    if expectedPassword == "" {
        log.Fatal("SFTP_PASSWORD environment variable not set")
    }
    return password == expectedPassword
}

func handleConnection(conn net.Conn) {
    defer conn.Close()
    privateBytes, err := os.ReadFile(privateKeyPath)
    if err != nil {
        log.Printf("Failed to read private key: %v", err)
        return
    }
    privateKey, err := ssh.ParsePrivateKey(privateBytes)
    if err != nil {
        log.Printf("Failed to parse private key: %v", err)
        return
    }

    config := &ssh.ServerConfig{
        PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
            if auth(c.User(), string(pass)) {
                return &ssh.Permissions{}, nil
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
    if newChannel.ChannelType() != "session" {
        newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
        return
    }

    channel, reqs, err := newChannel.Accept()
    if err != nil {
        log.Printf("Could not accept channel: %v", err)
        return
    }
    defer channel.Close()

    go ssh.DiscardRequests(reqs)
    server, err := sftp.NewServer(
        channel,
        sftp.WithRoot(rootDir),
    )
    if err != nil {
        log.Printf("Failed to start SFTP server: %v", err)
        return
    }
    defer server.Close()

    if err := server.Serve(); err != nil {
        log.Printf("SFTP server error: %v", err)
    }
}

func main() {
    absRootDir, err := filepath.Abs(rootDir)
    if err != nil {
        log.Fatalf("Failed to get absolute path: %v", err)
    }

    if err := os.MkdirAll(absRootDir, 0755); err != nil {
        log.Fatalf("Failed to create directory: %v", err)
    }

    if os.Getenv("SFTP_PASSWORD") == "" {
        log.Fatal("SFTP_PASSWORD environment variable not set")
    }
    if _, err := os.Stat(privateKeyPath); os.IsNotExist(err) {
        log.Fatalf("Private key file %s does not exist. Please generate one using: ssh-keygen -t rsa -b 2048 -f id_rsa -N \"\"", privateKeyPath)
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
                fmt.Printf("IP: %s%s\n", ipNet.IP.String(), port)
            }
        }
    }
    fmt.Printf("SFTP Server started on port %s\n", port)
    fmt.Printf("Files will be stored in: %s\n", absRootDir)

    for {
        conn, err := listener.Accept()
        if err != nil {
            log.Printf("Failed to accept connection: %v", err)
            continue
        }
        go handleConnection(conn)
    }
}
