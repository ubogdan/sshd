package svr

import (
	"fmt"
	"github.com/bytegang/sshd/svr/client"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"log"
)

func (s *Server) handleChannels(channels <-chan ssh.NewChannel, sshConn *ssh.ServerConn) {
	for someChannel := range channels {
		if t := someChannel.ChannelType(); t != "session" {
			err := someChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
			if err != nil {
				log.Println("ch.Reject err: ", err)
			}
			continue
		}
		// Multiplexing ssh control master
		go handleSingleChannelSession(someChannel, sshConn, s.Cfg.RpcAddr)
	}
}

func handleSingleChannelSession(ch ssh.NewChannel, sshConn *ssh.ServerConn, rpcAddr string) {
	user, err := getPermissionUser(sshConn)
	if err != nil {
		log.Println("the permission of auth has not user json", err)
		return
	}

	c, err := client.NewClient(user, sshConn, rpcAddr)
	if err != nil {
		log.Println("create ssh session client failed: ", err)
		return
	}
	defer c.Close()

	channel, requests, err := ch.Accept()
	if err != nil {
		log.Println("Could not accept channel: ", err)
		return
	}
	defer channel.Close()

	c.InitTerm(channel)
	channel.Write([]byte(fmt.Sprintf("Hello %s \r\n", user.Name)))
	for req := range requests {
		if req == nil {
			continue
		}
		switch req.Type {
		case "shell":
			go c.HandleShell(channel)
			req.Reply(true, []byte{})
			req.WantReply = false
		case "pty-req":
			c.ReqPtyPayload = req.Payload
			pl := new(payloadPtyReq)
			err := ssh.Unmarshal(req.Payload, pl)
			if err != nil {
				logrus.WithError(err).Error("pty-req")
			} else {
				c.SetWinSize(int(pl.Cw), int(pl.Ch))

			}
			err = req.Reply(true, nil)
			if err != nil {
				log.Println(err)
			}

		case "window-change":
			pl := new(payloadWinChange)
			err = ssh.Unmarshal(req.Payload, pl)
			if err != nil {
				log.Println(err)
			} else {
				c.Resize(int(pl.Cw), int(pl.Ch))

			}
			req.Reply(true, []byte{})
			req.WantReply = false

		case "exec":
			pl := new(payloadExecOrSubsystem)
			err := ssh.Unmarshal(req.Payload, pl)
			if err != nil {
				log.Printf("error parsing ssh execMsg: %s\n", err)
				return
			}
			go c.HandleExec(pl.CmdOrSubsystem, channel)
		case "env":
			pl := new(payloadEnv)
			err := ssh.Unmarshal(req.Payload, pl)
			if err != nil {
				log.Println(err)
			} else {
				c.ReqEnvPayload = req.Payload
			}
		case "subsystem":
			pl := new(payloadExecOrSubsystem)
			err := ssh.Unmarshal(req.Payload, pl)
			if err != nil {
				log.Println(err)
			} else {
				if pl.CmdOrSubsystem == "sftp" {
					go c.HandleSftp(channel)
				}
			}

		default:
			log.Println(req.Type, string(req.Payload))
		}

	}
}

//    boolean   want_reply
//      string    TERM environment variable value (e.g., vt100)
//      uint32    terminal width, characters (e.g., 80)
//      uint32    terminal height, rows (e.g., 24)
//      uint32    terminal width, pixels (e.g., 640)
//      uint32    terminal height, pixels (e.g., 480)
//      string    encoded terminal modes
type payloadPtyReq struct {
	Name string
	Cw   uint32
	Ch   uint32
	Pw   uint32
	Ph   uint32
	Mode string
}

//      byte      SSH_MSG_CHANNEL_REQUEST
//      uint32    recipient channel
//      string    "env"
//      boolean   want reply
//      string    variable name
//      string    variable value
type payloadEnv struct {
	EnvKey, EnvValue string
}

type payloadExecOrSubsystem struct {
	CmdOrSubsystem string
}

//      byte      SSH_MSG_CHANNEL_REQUEST
//      uint32    recipient channel
//      string    "window-change"
//      boolean   FALSE
//      uint32    terminal width, columns
//      uint32    terminal height, rows
//      uint32    terminal width, pixels
//      uint32    terminal height, pixels
type payloadWinChange struct {
	Cw uint32
	Ch uint32
	Pw uint32
	Ph uint32
}
