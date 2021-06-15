package sshweb

import (
	"context"
	"github.com/bytegang/pb"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"log"
	"net/http"
	"strconv"
	"time"
)

// wsTerminal websocket xterm
// @Summary websocket xterm
// @Description websocket xterm
// @Tags Websocket
// @Accept  json
// @Produce  json
// @Param t query string true "TOKEN of gRPC to exchange asset's ssh connection information"
// @Param r query integer true "terminal rows count"
// @Param c query integer true "terminal columns count"
// @Success 200 {object} responseObj
// @Router /api/ws/ssh [get]
func (ap *app) wsTerminal() func(c *gin.Context) {
	upGrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024 * 1024 * 10,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	return func(c *gin.Context) {
		w := c.Writer
		r := c.Request
		wsConn, err := upGrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		defer wsConn.Close()

		cols := r.FormValue("c")
		intC, err := strconv.Atoi(cols)
		if wsError(wsConn, err) {
			return
		}

		rows := r.FormValue("r")
		intR, err := strconv.Atoi(rows)
		if wsError(wsConn, err) {
			return
		}

		token := r.FormValue("t")

		sws, err := newWebsocketShell(token, ap.AddrRpc)
		if wsError(wsConn, err) {
			return
		}
		defer sws.Close()

		err = sws.runShell(intC, intR, wsConn)
		if wsError(wsConn, err) {
			return
		}
	}
}
func wsError(ws *websocket.Conn, err error) bool {
	if err != nil {
		logrus.WithError(err).Error("handler ws ERROR:")
		dt := time.Now().Add(time.Second)
		if err := ws.WriteControl(websocket.CloseMessage, []byte(err.Error()), dt); err != nil {
			logrus.WithError(err).Error("websocket writes control message failed:")
		}
		return true
	}
	return false
}

func newWebsocketShell(token, rpcAddr string) (ins *websocketSshShell, err error) {
	var opts = []grpc.DialOption{grpc.WithInsecure()}
	conn, err := grpc.Dial(rpcAddr, opts...)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	machineSshConfig, err := pb.NewByteGangsterClient(conn).WebXtermSsh(context.Background(), &pb.ReqToken{Token: token})
	if err != nil {
		return nil, err
	}
	return newWebsocketSessionShell(machineSshConfig, rpcAddr), nil
}
