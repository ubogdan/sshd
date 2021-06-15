package sshweb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/bytegang/pb"
	"github.com/bytegang/sshd/pkg/util"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
	"google.golang.org/grpc"
	"io"
	"log"
	"time"
)

const (
	timeStep     = time.Millisecond * time.Duration(10)
	timeStepWait = time.Millisecond * time.Duration(20)
)

type websocketSshShell struct {
	wsSshInfo     *pb.ResSshConnCfg
	SshClient     *ssh.Client
	SshSession    *ssh.Session
	stdInPipe     io.WriteCloser
	comboOutput   *safeBuffer //ssh 终端混合输出
	frames        []*pb.RecordFrame
	lineBuff      *bytes.Buffer
	lineTerm      *term.Terminal
	InputCommands []string
	AddrRpc       string
}

func newWebsocketSessionShell(wsSshInfo *pb.ResSshConnCfg, addrRpc string) *websocketSshShell {
	buff := new(bytes.Buffer)
	return &websocketSshShell{
		wsSshInfo:     wsSshInfo,
		SshClient:     nil,
		SshSession:    nil,
		stdInPipe:     nil,
		comboOutput:   new(safeBuffer),
		frames:        nil,
		lineBuff:      buff,
		lineTerm:      term.NewTerminal(buff, ""),
		InputCommands: []string{},
		AddrRpc:       addrRpc,
	}
}

func (hub *websocketSshShell) runShell(cols, rows int, wsConn *websocket.Conn) (err error) {
	err = hub.initSshSessionShell(cols, rows)
	if err != nil {
		return err
	}
	qChan := make(chan bool, 3)
	go hub.wsToSsh(wsConn, qChan)
	go hub.sshToWs(wsConn, qChan)
	go hub.wait(qChan)
	<-qChan

	return nil
}

func (hub *websocketSshShell) initSshSessionShell(cols, rows int) (err error) {
	sshClient, err := util.GetSshClient(hub.wsSshInfo.AssetConn, hub.wsSshInfo.ProxyConn, time.Second*3)
	if err != nil {
		return fmt.Errorf("登录SSH失败：ssh %s@%s   %v", hub.wsSshInfo.AssetConn.User, hub.wsSshInfo.AssetConn.Addr, err)
	}
	sshSession, err := sshClient.NewSession()
	if err != nil {
		return err
	}
	hub.SshSession = sshSession

	hub.stdInPipe, err = sshSession.StdinPipe()
	if err != nil {
		return err
	}
	//ssh.stdout and stderr will write output into comboWriter
	sshSession.Stdout = hub.comboOutput
	sshSession.Stderr = hub.comboOutput

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // disable echo
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}
	// Request pseudo terminal
	if err := sshSession.RequestPty("xterm", rows, cols, modes); err != nil {
		return err
	}
	// initSshSessionShell remote shell
	if err := sshSession.Shell(); err != nil {
		return err
	}
	return nil
}

//Close 关闭
func (hub *websocketSshShell) Close() {
	if hub.SshSession != nil {
		hub.SshSession.Close()
	}
	if hub.SshClient != nil {
		hub.SshClient.Close()
	}
	if hub.comboOutput != nil {
		hub.comboOutput = nil
	}

	//send logs to rpc
	addr := hub.AddrRpc
	var opts = []grpc.DialOption{grpc.WithInsecure()}
	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	req := pb.ReqSshdData{
		Uuid:   hub.wsSshInfo.Uuid,
		Frames: hub.frames,
	}

	_, err = pb.NewByteGangsterClient(conn).SaveLogSshSession(context.Background(), &req)
	if err != nil {
		log.Println(err)
		return
	}

}

func (hub *websocketSshShell) isMachineUserExpired() bool {
	//todo::检查授权过期
	return false
}

func (hub *websocketSshShell) isMatchCmdRule(line string) (stop bool) {
	//todo:: 命令拦截
	return false
}

var clearLine = []byte{27, '[', '2', 'K'}
var warning = []byte("您输入的是限制命令" + "\r")

func (hub *websocketSshShell) wsMsgToSsh(cmdBytes []byte, wsConn *websocket.Conn) {
	var err error
	//保存整行input
	var lineCommand string
	hub.lineTerm.Write(cmdBytes)
	for _, bb := range cmdBytes {
		//判断命令是否开始换行或者;
		var stop bool
		if bb == '\r' {
			lineCommand, err = hub.lineTerm.ReadLine()
			//log.Println(lineCommand)
			if err != nil {
				logrus.WithError(err).Error("read terminal line command failed")
			}
			hub.lineBuff.Reset()
			hub.InputCommands = append(hub.InputCommands, lineCommand)
			stop = hub.isMatchCmdRule(lineCommand)
			if stop {
				//1. clear previous
				//hub.stdInPipe.Write(clearLine)
				hub.stdInPipe.Write(makeBackSpace(lineCommand))
				time.Sleep(timeStepWait)
				wsConn.WriteMessage(websocket.TextMessage, clearLine)
				wsConn.WriteMessage(websocket.TextMessage, warning)
				ins := new(pb.RecordFrame)
				ins.Operation = pb.MsgOperation_Warning
				ins.Data = string(warning)
				hub.record(ins)
				//2. show warning msg
			}
		}
		// stop 则不会执行最后一次enter
		if !stop {
			_, err := hub.stdInPipe.Write([]byte{bb})
			if err != nil {
				log.Print(err)
			}
		}
	}
	//hub.writeLogs(cmdBytes)
}

const backspace = byte(127)

func makeBackSpace(line string) []byte {
	res := make([]byte, len(line))
	for i := range line {
		res[i] = backspace
	}
	return res
}

func (hub *websocketSshShell) sshToWs(wsConn *websocket.Conn, exitCh chan bool) {
	//tells other go routine quit
	defer setQuit(exitCh)

	//every 120ms write combine output bytes into websocket response
	tick := time.NewTicker(timeStep)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			if hub.comboOutput == nil {
				return
			}
			bs := hub.comboOutput.Bytes()
			if len(bs) > 0 {
				hub.comboOutput.Reset()
				msgPtr := &pb.RecordFrame{
					Operation: pb.MsgOperation_Stdout,
					Data:      string(bs),
				}
				dataBytes, err := json.Marshal(msgPtr)
				if err != nil {
					logrus.WithError(err).Error("ssh sending combo output to webSocket failed")
				}
				err = wsConn.WriteMessage(websocket.TextMessage, dataBytes)
				if err != nil {
					logrus.WithError(err).Error("ssh sending combo output to webSocket failed")
				}
				hub.record(msgPtr)
			}

		case <-exitCh:
			return
		}
	}
}

func (hub *websocketSshShell) wait(quitChan chan bool) {
	defer setQuit(quitChan)
	if err := hub.SshSession.Wait(); err != nil {
		logrus.WithError(err).Error("ssh session wait failed")
	}
}

func setQuit(ch chan bool) {
	ch <- true
}

func (hub *websocketSshShell) wsToSsh(wsConn *websocket.Conn, exitCh chan bool) {
	tick := time.NewTicker(time.Second * 180)
	defer tick.Stop()
	defer setQuit(exitCh)
	for {
		select {
		case <-tick.C:
			// check the machine is expired
			if hub.isMachineUserExpired() {
				return
			}
			// web点击断开 实时关闭链接
			//if sshd.BlockMgrHas(hub.UserID, hub.MachineUserID) {
			//	logrus.Info("real time block triggered")
			//	return
			//}
		case <-exitCh:
			return
		default:
			//read websocket msg
			_, wsData, err := wsConn.ReadMessage()
			if err != nil {
				logrus.WithError(err).Error("reading webSocket message failed")
				return
			}
			msgObj := new(pb.RecordFrame)
			if err := json.Unmarshal(wsData, msgObj); err != nil {
				logrus.WithError(err).WithField("wsData", string(wsData)).Error("unmarshal sshweb socket message failed")
			}
			switch msgObj.Operation {
			case pb.MsgOperation_Resize:
				//handle xterm.js size change
				if msgObj.Cols > 0 && msgObj.Rows > 0 {
					if err := hub.SshSession.WindowChange(int(msgObj.Rows), int(msgObj.Cols)); err != nil {
						logrus.WithError(err).Error("ssh pty change windows size failed")
					}
					hub.record(msgObj)
				}
			case pb.MsgOperation_Stdin:
				//handle xterm.js stdin
				hub.wsMsgToSsh([]byte(msgObj.Data), wsConn)
				hub.record(msgObj)

			case pb.MsgOperation_Ping:
				//hub.TouchAt = time.Now()

			}
		}
	}
}
