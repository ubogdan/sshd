package client

import (
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"strings"
)

func (c *Client) HandleShell(channel ssh.Channel) {
	myActions := getActionStoreBy(c.User.Role)
	defer func() {
		//err := c.postman.UserOffline(c.SshUser)
		//if err != nil {
		//	log.Println("user offline", err)
		//}
		channel.Close()
	}()
	//err := c.postman.UserOnline(c.SshUser)
	//if err != nil {
	//	log.Println("user offline", err)
	//	return
	//}

	//exitChan := make(chan bool, 1)
	// FIXME: This shouldn't live here, need to restructure the call chaining.
	//c.Server.Add(c)

	//err = c.postman.RegisterClientDevice(c.SshUser, c.DeviceSessionID)
	//if err != nil {
	//	logrus.Println(err)
	//	return
	//}
	go func() {
		// Block until done, then remove.
		c.Conn.Wait()
		c.closed = true
		//c.Server.Remove(c)
		//close(c.Messages)
		//c.postman.UserOffline(c.SshUser)
	}()

	go func() {
		//todo:: send history msg
		//for msg := range c.Messages {
		//	c.Write(msg)
		//}
		//todo:: 解决goroutine race问题
		//c.postman.ReceiveMsgLoop(c.DeviceSessionID, c, exitChan)
	}()
	new(actionHelp).Exec(c, nil) // 打印帮助信息
	for {
		if c.OnBastion {
			continue
		}
		line, err := c.term.ReadLine()
		if err != nil {
			break
		}
		var doer ActionDoer = new(actionHelp)
		// choose action
		isCmd, action, args := parseInputLine(line)
		if isCmd {
			v, ok := myActions[action]
			if ok {
				doer = v
			} else {
				c.Danger("无效命令请查看一下帮助说明: " + line)
				//continue
			}
		}
		//exec matched action
		if hint := doer.Hint(&args); hint != "" { //check arg
			c.Warning("Invalid command: " + line)
			continue
		}
		err = doer.Exec(c, args)
		if err != nil {
			c.Warning(err.Error())
			logrus.Error(err)
			//c.TermWrite(err.Error())
		}
		c.restorePrompt()
	}

}

const cmdPrefix = "/"

func parseInputLine(line string) (isCmd bool, action string, args []string) {
	parts := strings.Split(line, " ")
	if len(parts) > 0 {
		args = []string{}
		for _, p := range parts {
			if t := strings.TrimSpace(p); t != "" {
				args = append(args, t)
			}
		}
		return strings.HasPrefix(parts[0], cmdPrefix), strings.TrimPrefix(parts[0], cmdPrefix), args
	}
	return false, "", []string{line}
}
