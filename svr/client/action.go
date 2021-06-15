package client

import (
	pb "github.com/bytegang/pb"
	"log"
	"sort"
)

//ActionDoer 编写插件hook 来扩展更多的功能
type ActionDoer interface {
	Help() (cmd, alias, help string)
	Allow(role pb.UserRole) bool
	Exec(c *Client, args []string) error
	Hint(args *[]string) string
}

var allActions = map[string]ActionDoer{}

//registerAction 注册编写的action hook 扩展功能
func registerAction(do ActionDoer) {
	cmd, alias, _ := do.Help()

	if cmd == "" {
		log.Fatal("action cmd must not be empty string")
	}

	_, ok := allActions[cmd]
	if ok {
		log.Fatal("action cmd has already existed: ", cmd)
	} else {
		allActions[cmd] = do
	}

	if alias != "" {
		_, ok = allActions[alias]
		if ok {
			log.Fatal("action alias has already existed: ", alias)
		} else {
			allActions[alias] = do
		}
	}
	allHelpActions = append(allHelpActions, do)
	sort.Sort(allHelpActions)
}

var allHelpActions helpList

type helpList []ActionDoer

func (h helpList) Len() int {
	return len(h)
}

func (h helpList) Less(i, j int) bool {
	cmdi, _, _ := h[i].Help()
	cmdj, _, _ := h[j].Help()
	return cmdi < cmdj
}

func (h helpList) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}
