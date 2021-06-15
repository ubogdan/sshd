package client

import (
	"github.com/bytegang/pb"
)

var _ ActionDoer = new(ActionDefault)

type ActionDefault struct{}

func (a ActionDefault) Help() (cmd, alias, log string) {
	return "sys", "", "sys"
}
func (a ActionDefault) Allow(role pb.UserRole) bool {
	return true
}
func (a ActionDefault) Exec(c *Client, args []string) error {
	//msg := joinMsg(c, args)
	//if msg == "" {
	//	return nil
	//}
	//if c.selectedFriend != nil { //send a msg to a selected friend
	//	err := c.postman.SendMsgUser(c.SshUser, c.selectedFriend, msg)
	//	return err
	//
	//} else if group := c.selectedGroup; group != nil { //send a message to a group
	//	for _, u := range group.Members {
	//		err := c.postman.SendMsgUser(c.SshUser, &u, msg)
	//		if err != nil {
	//			logrus.Error(err)
	//		}
	//	}
	//	return nil
	//} else {
	//	return c.postman.SendMsgBroadCast(c.SshUser, msg)
	//}
	return nil
}

func (a ActionDefault) Hint(args *[]string) string {
	return ""
}
