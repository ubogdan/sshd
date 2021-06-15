package client

import (
	"context"
	"errors"
	"fmt"
	"github.com/bytegang/pb"
	"github.com/bytegang/sshd/pkg/util"
	"github.com/fatih/color"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
	"google.golang.org/grpc"
	"log"
	"time"
)

type Client struct {
	RpcCli          pb.ByteGangsterClient
	RpcConn         *grpc.ClientConn
	DeviceSessionID string
	SshChan         ssh.Channel
	Conn            *ssh.ServerConn
	User            *pb.User
	selectedFriend  *pb.User
	selectedGroup   *pb.Group
	Color           string
	IsAdmin         bool
	ready           chan struct{}
	term            *term.Terminal
	termWidth       int
	termHeight      int
	silencedUntil   time.Time
	LastTX          time.Time
	beepMe          bool
	colorMe         bool
	closed          bool
	OnBastion       bool
	WinSizeChan     chan WinSize
	ReqPtyPayload   []byte //bastion action need the payload
	ReqEnvPayload   []byte //bastion set env
}

// NewClient constructs a new client
// 1.记录client terminal的状态
// 2.当前用户的状态
// 3.消息发送接收
// 4.好友群组关系管理
// 5.读取客户段terminal的输入
func NewClient(user *pb.User, sshConn *ssh.ServerConn, rpcAddr string) (*Client, error) {
	var opts = []grpc.DialOption{grpc.WithInsecure()}
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*3)
	defer cancelFunc()
	rpcConn, err := grpc.DialContext(ctx, rpcAddr, opts...)
	if err != nil {
		return nil, fmt.Errorf("rpc-service is not available: %s, err: %v", rpcAddr, err)
	}
	rpcClient := pb.NewByteGangsterClient(rpcConn)

	return &Client{
		RpcCli:          rpcClient,
		RpcConn:         rpcConn, //todo need to close
		DeviceSessionID: string(sshConn.SessionID()),
		SshChan:         nil,
		Conn:            sshConn,
		User:            user,
		selectedFriend:  nil,
		selectedGroup:   nil,
		Color:           util.RandomColor256(),
		IsAdmin:         false,
		ready:           make(chan struct{}, 1),
		term:            nil,
		termWidth:       0,
		termHeight:      0,
		silencedUntil:   time.Time{},
		LastTX:          time.Now(),
		beepMe:          false,
		colorMe:         true,
		closed:          false,
		OnBastion:       false,
		WinSizeChan:     make(chan WinSize),
		ReqPtyPayload:   nil,
		ReqEnvPayload:   nil,
	}, nil
}

func (c *Client) InitTerm(rw ssh.Channel) {
	c.term = term.NewTerminal(rw, "")
	c.restorePrompt()
	c.SshChan = rw
}

func (c *Client) SetPrompt(s string) {
	c.term.SetPrompt(util.ColorInvert + util.ColorBold + c.Color + fmt.Sprintf(" %s ", s) + util.ColorReset + " ")
}

func (c *Client) Danger(msg string) {
	//🔴 Red Circle
	//🟠 Orange Circle
	//🟡 Yellow Circle
	//🟢 Green Circle
	//🔵 Blue Circle
	//🟣 Purple Circle
	//🟤 Brown Circle
	//⚫ Black Circle
	//⚪ White Circle
	content := color.RedString("🔴%s\r\n", msg)
	c.writePlain(content)
}
func (c *Client) Warning(msg string) {
	content := color.YellowString("🟠%s\r\n", msg)
	c.writePlain(content)
}

func (c *Client) Success(msg string) {
	content := color.GreenString("🟢%s\r\n", msg)
	c.writePlain(content)
}

func (c *Client) Primary(msg string) {
	content := color.BlueString("🔵%s\r\n", msg)
	c.writePlain(content)
}

func (c *Client) MsgPrivate(msg string) {
	content := color.HiCyanString("️%s\r\n", msg)
	c.writePlain(content)
}

func (c *Client) MsgGroup(msg string) {
	content := color.HiYellowString("%s\r\n", msg)
	c.writePlain(content)
}

func (c *Client) TermQA(questions []string, echos []bool) (answers []string, err error) {
	if len(questions) != len(echos) {
		return nil, errors.New("questions must match echos")
	}
	answers = make([]string, len(questions))
	for idx, q := range questions {
		c.term.SetPrompt(q)
		if echos[idx] {
			line, err := c.term.ReadLine()
			if err != nil {
				return nil, err
			}
			answers[idx] = line
		} else {
			line, err := c.term.ReadPassword(q)
			if err != nil {
				return nil, err
			}
			answers[idx] = line
		}
	}
	return answers, nil
}

func (c *Client) Close() {
	if c.RpcConn != nil {
		err := c.RpcConn.Close()
		if err != nil {
			log.Println(err)
		}
	}
}
