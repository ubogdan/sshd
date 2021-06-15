package client

import (
	"context"
	"errors"
	"fmt"
	"github.com/bytegang/pb"
	"github.com/bytegang/sshd/pkg/util"
	"github.com/olekukonko/tablewriter"
	"golang.org/x/term"
	"io"
	"log"
	"strconv"
	"strings"
	"time"
)

func init() {
	registerAction(new(actionBastion))
}

type actionBastion struct{}

func (a actionBastion) Help() (cmd, short, log string) {
	return "bastion", "b", "堡垒机"
}
func (a actionBastion) Allow(role pb.UserRole) bool {
	return role == pb.UserRole_Developer
}

func showTable(w io.Writer, data [][]string, headers []string) {
	table := tablewriter.NewWriter(w)
	table.SetRowLine(true)
	table.SetHeader(headers)
	for _, v := range data {
		table.Append(v)
	}
	table.Render()
}

func assetChoose(c io.ReadWriter, rows []*pb.Asset) (a *pb.Asset, err error) {
	listPrompt := "asset list:"
	retry := 3
	selectPrompt := "please input the asset index number of you want:"

	terminal := term.NewTerminal(c, listPrompt)
	var data [][]string
	for idx, row := range rows {
		id := fmt.Sprintf("%d", idx+1)
		data = append(data, []string{id, row.Hostname, row.ShhAddr, row.Alias, row.Remark})
	}
	showTable(terminal, data, []string{"ID", "Hostname", "Addr", "Alias", "Remark"})

	for i := 0; i < retry; i++ {
		selectPrompt = "Please input the row index you want:"
		terminal.SetPrompt(selectPrompt)
		line, err := terminal.ReadLine()
		if err != nil {
			return nil, err
		}
		line = strings.TrimSpace(line)
		if line == "q" || line == "exit" || line == "quite" {
			return nil, errors.New("user exit the selection")
		}
		parseInt, err := strconv.ParseInt(line, 10, 64)
		if err != nil {
			terminal.Write([]byte("your input is not a valid index number,please retry"))
			continue
		}
		if parseInt < 1 || int(parseInt) > len(rows) {
			terminal.Write([]byte("your input number is out of range,please retry"))
			continue
		}
		return rows[int(parseInt-1)], nil
	}
	return nil, errors.New("sorry, you run out of retries")
}

func (a actionBastion) Exec(c *Client, args []string) error {
	c.OnBastion = true
	defer func() {
		c.OnBastion = false
	}()
	rpcClient := c.RpcCli

	in := pb.ReqAssetsQuery{
		UserId: c.User.Id,
		Query:  "",
	}
	if len(args) > 1 {
		in.Query = strings.TrimSpace(args[1])
	}

	assetList, err := rpcClient.FetchAsset(context.Background(), &in)
	if err != nil {
		return err
	}
	//render asset list for selection
	chosenAsset, err := assetChoose(c.SshChan, assetList.List)
	if err != nil {
		return err
	}

	//fetch ssh user
	sshAccount, err := rpcClient.FetchAssetSshConfig(context.Background(), &pb.ReqAssetUser{
		AssetId: chosenAsset.Id,
		UserId:  c.User.Id,
	})
	if err != nil {
		return err
	}

	client, err := util.GetSshClient(sshAccount.AssetConn, sshAccount.ProxyConn, time.Second*10)
	if err != nil {
		return err
	}

	bastionSess, err := NewBastionSession(client, sshAccount.AssetConn.Addr, c.ReqPtyPayload, c.ReqEnvPayload, c.SshChan, c.termHeight, c.termWidth)
	if err != nil {
		return err
	}
	defer func() {
		err := bastionSess.Close()
		if err != nil {
			log.Println("bastion Sess.Close", err)
		}
	}()
	err = bastionSess.Start(c.SshChan, c.WinSizeChan)
	if err != nil {
		log.Println("start-", err)
	}
	return nil
}

func (a actionBastion) Hint(args *[]string) string {

	// 重点关注文件
	// code/koko/pkg/proxy/commonswitch.go
	// WrapperSession
	//type ServerSSHConnection struct {
	//	session *gossh.Session
	//	stdin   io.Writer
	//	stdout  io.Reader
	//	options *connectionOptions
	//}
	return ""
}
