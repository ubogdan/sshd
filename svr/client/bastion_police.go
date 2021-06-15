package client

import (
	"bytes"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
	"log"
	"sync"
	"time"
)

// ssh.marshal
type item struct {
	UnixNano int64  `json:"u"`
	Width    int    `json:"w"`
	Height   int    `json:"h"`
	Data     []byte `json:"d"`
}

type Rec struct {
	Commands []string `json:"c"`
	Lines    []item   `json:"l"`
}

type CmdEvent struct {
	Lvl uint
	Msg string
}

// 记录输入输出和检测敏感输入
type PoliceRecorder struct {
	sync.Mutex
	termH    int
	termW    int
	userChan ssh.Channel    // send msg of guardian to user shell terminal
	term     *term.Terminal // parse input command
	camera   Rec
	events   chan CmdEvent
}

func NewPoliceRecorder(us ssh.Channel, w, h int) *PoliceRecorder {
	wr := bytes.NewBuffer(nil)
	t := term.NewTerminal(wr, "")
	err := t.SetSize(w, h)
	if err != nil {
		log.Println(err)
	}
	return &PoliceRecorder{
		Mutex:    sync.Mutex{},
		termH:    h,
		termW:    w,
		userChan: us,
		term:     t,
		camera:   Rec{},
		events:   make(chan CmdEvent),
	}
}

func (p *PoliceRecorder) WriteStdin(buf []byte) (n int, err error) {
	p.Lock()
	defer p.Unlock()

	wn, err := p.term.Write(buf)
	if err != nil {
		return 0, err
	}
	for i := range buf {
		if buf[i] == '\r' || buf[i] == '\n' {
			line, err := p.term.ReadLine()
			if err != nil {
				log.Println(err)
			} else {
				p.ruleMatch(line)
			}
		}
	}
	//记录日志
	p.camera.Lines = append(p.camera.Lines, p.makeItem(buf))
	return wn, nil
}

func (p *PoliceRecorder) WriteStdout(buf []byte) (n int, err error) {
	p.Lock()
	defer p.Unlock()

	//wn, err := p.term.Write(buf)
	//if err != nil {
	//	return 0, err
	//}

	//记录日志
	p.camera.Lines = append(p.camera.Lines, p.makeItem(buf))
	return len(buf), nil
}

func (p *PoliceRecorder) Close() error {
	close(p.events)
	//todo:: 这一部分工作 给grpc svr做
	//开始保持日志
	//bs, err := json.Marshal(p.camera)
	//if err != nil {
	//	return err
	//}
	////自定义文件名称
	//err = os.WriteFile("ssh.video.js", bs, os.ModePerm)
	//if err != nil {
	//	return err
	//}
	return nil
}

func (p *PoliceRecorder) ruleMatch(cmd string) {
	p.camera.Commands = append(p.camera.Commands, cmd)
	//todo:: 匹配告警规则

}

func (p *PoliceRecorder) Resize(w, h int) error {
	p.termH = h
	p.termW = w
	return p.term.SetSize(w, h)
}

func (p *PoliceRecorder) makeItem(data []byte) item {
	return item{
		UnixNano: time.Now().UnixNano(),
		Width:    p.termW,
		Height:   p.termH,
		Data:     data,
	}
}
