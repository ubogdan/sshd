package client

import pb "github.com/bytegang/pb"

func init() {
	registerAction(new(actionExit))
}

type actionExit struct{}

func (a actionExit) Help() (cmd, short, log string) {
	return "quit", "q", "关闭会话"
}
func (a actionExit) Allow(role pb.UserRole) bool {
	return true
}
func (a actionExit) Exec(c *Client, args []string) error {
	return c.SshChan.Close()
}

func (a actionExit) Hint(args *[]string) string {
	//if len(args) < 2 {
	//	return "Missing $THEME from: /theme $THEME" + " Choose either color or mono"
	//}
	return ""
}
