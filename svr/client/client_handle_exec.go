package client

import (
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type execMsg struct {
	Command string
}

type exitStatusMsg struct {
	Status uint32
}

//https://stackoverflow.com/questions/33846959/golang-ssh-server-how-to-handle-file-transfer-with-scp
func (c *Client) HandleExec(line string, ch ssh.Channel) {
	myActions := getActionStoreBy(c.User.Role)

	var doer ActionDoer = new(ActionDefault)
	// choose action
	_, action, args := parseInputLine(line)
	v, ok := myActions[action]
	if ok {
		doer = v
	} else {
		c.Danger("Invalid command: " + line)
		return
	}
	//exec matched action
	if hint := doer.Hint(&args); hint != "" { //check arg
		c.Danger("Invalid command: " + line)
		return
	}
	err := doer.Exec(c, args)
	if err != nil {
		c.Warning(err.Error())
		logrus.Error(err)
		//c.TermWrite(err.Error())
	}

	// ch can be used as a ReadWriteCloser if there should be interactivity
	ex := exitStatusMsg{
		Status: 0,
	}

	// return the status code
	if _, err := ch.SendRequest("exit-status", false, ssh.Marshal(&ex)); err != nil {
		logrus.Printf("unable to send status: %v", err)
	}
	ch.Close()
}

func (c Client) ParseCommandLine(req *ssh.Request) (string, error) {
	var msg execMsg
	if err := ssh.Unmarshal(req.Payload, &msg); err != nil {
		return "", err
	} else {
		return msg.Command, err
	}
}
