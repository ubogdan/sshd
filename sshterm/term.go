package sshterm

import (
	"errors"
	"fmt"
	"github.com/bytegang/pb"
	"github.com/bytegang/sshd/pkg/util"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
	"io"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

func StartTerm(target, proxy *pb.SshConn, timeout time.Duration) error {
	client, err := util.GetSshClient(target, proxy, timeout)
	if err != nil {
		return err
	}
	defer client.Close()
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	exitMsg := ""
	defer func() {
		if exitMsg == "" {
			fmt.Fprintln(os.Stdout, "[Felix]: the connection was closed on the remote side on ", time.Now().Format(time.RFC822))
		} else {
			fmt.Fprintln(os.Stdout, exitMsg)
		}
	}()
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		osName := runtime.GOOS
		return fmt.Errorf("%s fd %d is not a term, can't create pty of ssh", osName, fd)
	}
	state, err := term.MakeRaw(fd)
	if err != nil {
		return err
	}
	defer term.Restore(fd, state)

	termWidth, termHeight, err := term.GetSize(fd)
	if err != nil {
		return err
	}
	termType := os.Getenv("TERM")
	if termType == "" {
		termType = "xterm-256color"
	}

	err = session.RequestPty(termType, termHeight, termWidth, ssh.TerminalModes{})
	if err != nil {
		return err
	}

	pipStdIn, err := session.StdinPipe()
	if err != nil {
		return err
	}
	pipStdout, err := session.StdoutPipe()
	if err != nil {
		return err
	}
	pipStderr, err := session.StderrPipe()
	if err != nil {
		return err
	}
	quit := make(chan bool, 4)

	// todo:: 监听 quit 参数,
	// 其实不监听也可以, 因为这个方式是被命令行调用, 命令行quit 其他的goroutine都会自动quit
	go termResizeLoop(session)
	go io.Copy(os.Stderr, pipStderr)
	go io.Copy(os.Stdout, pipStdout)
	go io.Copy(pipStdIn, os.Stdin)

	err = session.Shell()
	if err != nil {
		quit <- true
		return err
	}
	err = session.Wait()
	if err != nil {
		quit <- true
		return err
	}
	return nil

}

func termResizeLoop(s *ssh.Session) error {

	// SIGWINCH is sent to the process when the window size of the terminal has
	// changed.
	sigwinchCh := make(chan os.Signal, 1)
	signal.Notify(sigwinchCh, syscall.SIGWINCH)

	fd := int(os.Stdout.Fd()) //
	termWidth, termHeight, err := term.GetSize(fd)
	if err != nil {
		return err
	}
	for sigwinch := range sigwinchCh {
		if sigwinch == nil {
			return errors.New("syscall.SIGWINCH signal chan is nil")
		}
		currTermWidth, currTermHeight, err := term.GetSize(fd)
		// Terminal size has not changed, don't do anything.
		if currTermHeight == termHeight && currTermWidth == termWidth {
			continue
		}

		err = s.WindowChange(currTermHeight, currTermWidth)
		if err != nil {
			fmt.Printf("Unable to send window-change reqest: %s.", err)
			continue
		}
		termWidth, termHeight = currTermWidth, currTermHeight
	}

	return nil
}
