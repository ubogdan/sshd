package client

import (
	"fmt"
	"github.com/bytegang/pb"
	"github.com/fatih/color"
	"strings"
)

const banner = `
@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
@      欢迎来到Eric Zhou的命令行SSH聊天交友灌水平台       @
@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
操作指南 更多相关信息请访问 https://mojotv.cn

`

func init() {
	registerAction(new(actionHelp))
}

type actionHelp struct{}

func (a actionHelp) Help() (cmd, short, log string) {
	return "help", "h", "help"
}

func (a actionHelp) Exec(c *Client, args []string) error {

	//colorReset := string([]byte{byte(27), '[', '0', 'm'})

	padCmd, padAlias := 0, 0
	for _, v := range allHelpActions {
		cmd, alias, _ := v.Help()
		if padCmd < len(cmd) {
			padCmd = len(cmd)
		}
		if padAlias < len(alias) {
			padAlias = len(alias)
		}
	}

	msg := banner
	for _, v := range getActionHelpBy(c.User.Role) {
		cmd, alias, commandDescribe := v.Help()

		spacedCommand := []rune(strings.Repeat(" ", padCmd))
		for idx, ss := range cmd {
			spacedCommand[idx] = ss
		}
		spacedAlias := []rune(strings.Repeat(" ", padAlias))
		for idx, ss := range alias {
			spacedAlias[idx] = ss
		}
		aliasString := string(spacedAlias)
		if alias == "" {
			aliasString = " " + aliasString
		} else {
			aliasString = cmdPrefix + aliasString
		}
		redCommand := color.New(color.FgRed, color.Bold).Sprintf("%s%s", cmdPrefix, string(spacedCommand))
		cyanAlias := color.New(color.Italic, color.FgCyan).Sprintf("%s", aliasString)
		commandDescribe = color.New(color.Faint).Sprintf("%s", commandDescribe)
		msg += fmt.Sprintf("%s %s %s\r\n", redCommand, cyanAlias, commandDescribe)
	}
	c.writePlain(msg)
	return nil
}

func (a actionHelp) Hint(args *[]string) string {
	return ""
}
func (a actionHelp) Allow(role pb.UserRole) bool {
	return true
}

func getActionStoreBy(role pb.UserRole) map[string]ActionDoer {
	//todo:: fix this error
	res := make(map[string]ActionDoer)
	for kk, do := range allActions {
		if do.Allow(role) {
			res[kk] = do
		}
	}
	return res
}

func getActionHelpBy(role pb.UserRole) (l helpList) {
	for _, vv := range allHelpActions {
		if vv.Allow(role) {
			l = append(l, vv)
		}
	}
	return l
}
