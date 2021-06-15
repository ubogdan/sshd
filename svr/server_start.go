package svr

import (
	"github.com/bytegang/sshd/pkg/util"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"log"
	"net"
	"syscall"
)

// Start starts the server
func (s *Server) Start(sshAddr string) error {
	// Once a ServerConfig has been configured, connections can be
	// accepted.
	socket, err := net.Listen("tcp", sshAddr)
	if err != nil {
		return err
	}
	log.Println("ssh-server Listening on", sshAddr)

	go func() {
		defer socket.Close()
		for {
			conn, err := socket.Accept()
			if err != nil {
				log.Println("failed to accept connection: ", err)
				if err == syscall.EINVAL {
					// TODO: Handle shutdown more gracefully?
					return
				}
			}

			// Goroutineify to resume accepting sockets early.
			go func() {
				// From a standard TCP connection to an encrypted SSH connection
				sshConn, channels, requests, err := ssh.NewServerConn(conn, s.sshConfig)
				if err != nil {
					log.Println("Failed to handshake: ", err)
					return
				}

				version := util.NormalString(string(sshConn.ClientVersion())) //reStripText.ReplaceAllString(string(sshConn.ClientVersion()), "")
				if len(version) > 100 {
					version = "evil puppy with a super long string"
				}
				logrus.Infof("Connection #%d from: %s, %s, %s", s.count+1, sshConn.RemoteAddr(), sshConn.User(), version)

				go ssh.DiscardRequests(requests)

				go s.handleChannels(channels, sshConn)
			}()
		}
	}()

	go func() {
		<-s.done
		socket.Close()
	}()
	return nil
}
