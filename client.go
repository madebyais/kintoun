package main

import (
	"fmt"
	"net"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// NewClient returns Client object
func NewClient(config *Config) *Client {
	return &Client{
		Config: config,
	}
}

// Client represents sftp and http client
type Client struct {
	Config *Config
}

// InitSftp initializes sftp connection
func (c *Client) InitSftp() *sftp.Client {
	sshconfig := &ssh.ClientConfig{
		User: c.Config.Source.Username,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		Auth: []ssh.AuthMethod{
			ssh.Password(c.Config.Source.Password),
		},
	}

	hostAddr := fmt.Sprintf("%s:%s", c.Config.Source.Host, c.Config.Source.Port)

	sshclient, errSSHClient := ssh.Dial("tcp", hostAddr, sshconfig)
	if errSSHClient != nil {
		panic(errSSHClient)
	}

	sftpclient, errSftpClient := sftp.NewClient(sshclient)
	if errSftpClient != nil {
		panic(errSftpClient)
	}

	return sftpclient
}
