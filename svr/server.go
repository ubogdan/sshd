package svr

import (
	"golang.org/x/crypto/ssh"
	"log"
	"time"
)

type Server struct {
	Cfg       *Config
	sshConfig *ssh.ServerConfig
	done      chan struct{}
	count     int
	motd      string
	whitelist map[string]struct{}   // fingerprint lookup
	admins    map[string]struct{}   // fingerprint lookup
	bannedPK  map[string]*time.Time // fingerprint lookup
	started   time.Time
}

// NewServer constructs a new server
func NewServer(cfg *Config) (*Server, error) {
	signer, err := ssh.ParsePrivateKey([]byte(cfg.PrivateKey))
	if err != nil {
		return nil, err
	}

	server := Server{
		Cfg:       cfg,
		sshConfig: nil,
		done:      make(chan struct{}),
		count:     0,
		motd:      "a pure ssh terminal chat server",
		whitelist: map[string]struct{}{},
		admins:    map[string]struct{}{},
		bannedPK:  map[string]*time.Time{},
		started:   time.Now(),
	}

	sshdCfg := ssh.ServerConfig{
		NoClientAuth: false,
		// Auth-related things should be constant-time to avoid timing attacks.
		PublicKeyCallback:           server.authPublicKeys(),
		KeyboardInteractiveCallback: server.authKeyboard(),
		PasswordCallback:            server.authPassword(),
		AuthLogCallback: func(conn ssh.ConnMetadata, method string, err error) {
			if err != nil {
				log.Println(conn, method, err)
			}
		},
	}
	sshdCfg.AddHostKey(signer)

	server.sshConfig = &sshdCfg
	return &server, nil
}
