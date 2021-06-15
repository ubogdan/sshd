package client

import (
	"fmt"
	"github.com/bytegang/pb"
)

var gitHash, buildAt string // build script set the value

func init() {
	registerAction(new(actionVersion))
}

type actionVersion struct{}

func (a actionVersion) Help() (cmd, alias, log string) {
	return "version", "v", "显示版本编译信息"
}
func (a actionVersion) Exec(c *Client, args []string) error {
	c.Primary(fmt.Sprintf("GitHash:%s 编译时间:%s", gitHash, buildAt))
	return nil
}

func (a actionVersion) Hint(args *[]string) string {
	return ""
}
func (a actionVersion) Allow(role pb.UserRole) bool {
	return true
}
