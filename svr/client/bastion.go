package client

import (
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"time"
)

type bastionSession struct {
	sshAddr      string
	sshConfig    *ssh.ClientConfig
	sshClient    *ssh.Client
	sess         *ssh.Session
	pipeOE       io.Reader //std-out std-err io.MultiReader
	pipeIn       io.WriteCloser
	PoliceOffice *PoliceRecorder
	exitChan     chan bool
}

func NewBastionSession(client *ssh.Client, sshAddr string, payloadPty []byte, payloadEnv []byte, us ssh.Channel, h, w int) (*bastionSession, error) {

	if len(payloadPty) == 0 {
		return nil, errors.New("payloadPty must not be nil")
	}

	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}

	//1. pty-req
	ok, err := session.SendRequest("pty-req", true, payloadPty)
	if err == nil && !ok {
		err = errors.New("ssh: pty-req failed")
	}
	if err != nil {
		return nil, err
	}

	//2. set env
	if len(payloadEnv) > 0 {
		ok, err := session.SendRequest("env", true, payloadEnv)
		if err == nil && !ok {
			err = errors.New("ssh: setenv failed")
		}
		if err != nil {
			return nil, err
		}
	}

	//3. set up io pipe line
	pipeIn, err := session.StdinPipe()
	if err != nil {
		return nil, err
	}
	pipeO, err := session.StdoutPipe()
	if err != nil {
		return nil, err
	}
	pipeE, err := session.StderrPipe()
	if err != nil {
		return nil, err
	}
	pipeOE := io.MultiReader(pipeO, pipeE)

	if err := session.Shell(); err != nil {
		return nil, err
	}

	return &bastionSession{
		sshAddr:      sshAddr,
		sshClient:    client,
		sess:         session,
		pipeOE:       pipeOE,
		pipeIn:       pipeIn,
		PoliceOffice: NewPoliceRecorder(us, w, h),
		exitChan:     make(chan bool, 1),
	}, nil
}

func (bs *bastionSession) Start(userRW ssh.Channel, winSizeChan <-chan WinSize) error {
	exitFn := func(name string) {
		if name != "" {
			log.Println("defer-exit-chan:", name)
		}
		bs.exitChan <- true
		time.Sleep(time.Second * 1)
	}

	// 1. watch user shell terminal windows size change
	go func() {
		defer exitFn("win-size")
		for {
			select {
			case ws := <-winSizeChan:
				err := bs.WinResize(ws.H, ws.W)
				if err != nil {
					log.Println("asset ssh session windows size change failed:", err)
					return
				}
			case <-bs.exitChan:
				log.Println("work-ws")

				return
			}
		}
	}()

	//2. proxy asset shell output to user ssh session
	go func() {
		defer exitFn("u->b")
		for {
			select {
			case <-bs.exitChan:
				//todo:: not work
				log.Println("work-ub")
				return
			default:
				_, err := lowCopy(bs, userRW)

				if err != nil && !errors.Is(err, io.EOF) {
					log.Println("u->b", err)
				}

			}
		}

	}()

	//3. proxy user ssh session std-in to asset shell
	go func() {
		defer exitFn("b->u")
		for {
			select {
			case <-bs.exitChan:
				log.Println("work-bs")
				return
			default:
				_, err := lowCopy(userRW, bs)
				if err != nil && !errors.Is(err, io.EOF) {
					log.Println("b->u", err)
				}
			}
		}
	}()

	//4. session-wait
	go func() {
		defer exitFn("session-wait")
		err := bs.sess.Wait()
		if err != nil && !errors.Is(err, io.EOF) {
			log.Println("wait-err", err)
		}
	}()

	//defer exitFn("session-wait")
	//err := bs.sess.wait()
	//if err != nil && !errors.Is(err, io.EOF) {
	//	log.Println("wait-err", err)
	//}
	defer exitFn("done")
	<-bs.exitChan

	return nil
}

func lowCopy(dst io.Writer, src io.Reader) (written int, err error) {
	buf := make([]byte, 1024)
	rn, err := src.Read(buf)
	if err != nil {
		return 0, err
	}
	return dst.Write(buf[:rn])
}

func (bs *bastionSession) Read(p []byte) (n int, err error) {
	rn, err := bs.pipeOE.Read(p)
	if err != nil {
		return 0, err
	}
	if rn > 0 {
		//save std-out to police
		_, err := bs.PoliceOffice.WriteStdout(p[:rn])
		if err != nil {
			return 0, fmt.Errorf("write police failed:%v", err)
		}
	}
	return rn, err
}

func (bs *bastionSession) Write(p []byte) (n int, err error) {
	wn, err := bs.pipeIn.Write(p)
	if err != nil {
		return 0, err
	}
	//save asset std-out to
	_, err = bs.PoliceOffice.WriteStdin(p)
	if err != nil {
		return 0, fmt.Errorf("write police failed:%v", err)
	}
	return wn, nil
}

func (bs *bastionSession) Close() (err error) {
	//log.Print("start close")
	if bs.pipeIn != nil {
		err := bs.pipeIn.Close()
		if err != nil && !errors.Is(err, io.EOF) {
			log.Println("pip in", err)
		}
	}
	if bs.PoliceOffice != nil {
		err := bs.PoliceOffice.Close()
		if err != nil {
			log.Println(err)
		}
	}
	if bs.sess != nil {
		err := bs.sess.Close()
		if err != nil && !errors.Is(err, io.EOF) {
			log.Println("session", err)
		}
	}
	if bs.sshClient != nil {
		err := bs.sshClient.Close()
		if err != nil {
			log.Println("sshClient close ", err)
		}
	}

	//close(bs.exitChan)
	return err
}

func (bs *bastionSession) WinResize(h, w int) error {
	err := bs.sess.WindowChange(h, w)
	if err != nil {
		return err
	}
	return bs.PoliceOffice.Resize(w, h)
}

type WinSize struct {
	H int
	W int
}
