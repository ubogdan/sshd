package sshweb

import (
	"bytes"
	"github.com/bytegang/pb"
	"io"
	"sync"
	"time"
)

func (hub *websocketSshShell) record(frame *pb.RecordFrame) {
	frame.Ts = uint64(time.Now().Unix())
	if frame.Operation == pb.MsgOperation_Ping {
		return
	}
	if frame.Operation == pb.MsgOperation_Resize && (frame.Rows < 1 || frame.Cols < 1) {
		return
	}
	hub.frames = append(hub.frames, frame)
}

type safeBuffer struct {
	buffer bytes.Buffer
	mu     sync.Mutex
}

func (w *safeBuffer) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buffer.Write(p)
}
func (w *safeBuffer) WriteTo(iw io.Writer) (int64, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buffer.WriteTo(iw)
}
func (w *safeBuffer) Bytes() []byte {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buffer.Bytes()
}
func (w *safeBuffer) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.buffer.Reset()
}
