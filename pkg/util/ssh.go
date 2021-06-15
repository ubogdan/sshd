package util

import (
	"context"
	"errors"
	"fmt"
	"github.com/bytegang/pb"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"time"
)

func publicKeyAuthFunc(pemBytes, keyPassword []byte) ssh.AuthMethod {
	if len(keyPassword) == 0 {
		signer, err := ssh.ParsePrivateKey(pemBytes)
		if err != nil {
			logrus.Println(err)
			return nil
		}
		return ssh.PublicKeys(signer)
	}

	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKeyWithPassphrase(pemBytes, keyPassword)
	if err != nil {
		logrus.WithError(err).Error("parse ssh key from bytes failed")
		return nil
	}
	return ssh.PublicKeys(signer)
}

func SshRemoteRunCommandWithTimeout(sshClient *ssh.Client, command string, timeout time.Duration) (string, error) {
	if timeout < 1 {
		return "", errors.New("timeout must be valid")
	}

	session, err := sshClient.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	ctx, cancelFunc := context.WithTimeout(context.Background(), timeout)
	defer cancelFunc()
	resChan := make(chan []byte, 1)
	errChan := make(chan error, 1)

	go func() {
		// run shell script
		if output, err := session.CombinedOutput(command); err != nil {
			errChan <- err
		} else {
			resChan <- output
		}
	}()

	select {
	case err := <-errChan:
		return "", err
	case result := <-resChan:
		return string(result), nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func NewSshClientConfig(sshUser, sshPassword, sshPrivateKey, sshPrivateKeyPassword string, timeout time.Duration) (config *ssh.ClientConfig, err error) {
	if sshUser == "" {
		return nil, errors.New("ssh_user can not be empty")
	}
	config = &ssh.ClientConfig{
		Timeout:         timeout,
		User:            sshUser,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //这个可以， 但是不够安全
		//HostKeyCallback: hostKeyCallBackFunc(h.Host),
	}
	if len(sshPrivateKey) == 0 {
		config.Auth = []ssh.AuthMethod{ssh.Password(sshPassword)}
	} else {
		config.Auth = []ssh.AuthMethod{publicKeyAuthFunc([]byte(sshPrivateKey), []byte(sshPrivateKeyPassword))}
	}
	return
}

func GetSshClient(target, proxy *pb.SshConn, timeout time.Duration) (c *ssh.Client, err error) {
	if target == nil {
		return nil, errors.New("target must not be nil")
	}
	targetConfig, err := NewSshClientConfig(target.User, target.Password, target.PrivateKey, target.PrivateKeyPassword, timeout)
	if err != nil {
		return nil, fmt.Errorf("target ssh config failed:%s", err)
	}
	if proxy == nil {
		return ssh.Dial("tcp", target.Addr, targetConfig)
	}

	//使用私有云集群跳板登陆
	proxyConfig, err := NewSshClientConfig(proxy.User, proxy.Password, proxy.PrivateKey, proxy.PrivateKeyPassword, timeout)
	if err != nil {
		return nil, fmt.Errorf("proxy ssh config failed:%s", err)
	}

	return createSshProxySshClient(targetConfig, proxyConfig, target.Addr, proxy.Addr)
}

func createSshProxySshClient(targetSshConfig, proxySshConfig *ssh.ClientConfig, targetAddr, proxyAddr string) (client *ssh.Client, err error) {
	proxyClient, err := ssh.Dial("tcp", proxyAddr, proxySshConfig)
	if err != nil {
		return
	}
	conn, err := proxyClient.Dial("tcp", targetAddr)
	if err != nil {
		return
	}
	ncc, chans, reqs, err := ssh.NewClientConn(conn, targetAddr, targetSshConfig)
	if err != nil {
		return
	}
	client = ssh.NewClient(ncc, chans, reqs)
	return
}
