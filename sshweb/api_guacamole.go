package sshweb

import (
	"bytes"
	"context"
	"fmt"
	"github.com/bytegang/pb"
	"github.com/bytegang/sshd/pkg/guac"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"log"
	"net/http"
)

func fetchGuacamoleArgFromRpc(token, rpcAddr string) (*pb.ResGuacamole, error) {
	var opts = []grpc.DialOption{grpc.WithInsecure()}
	conn, err := grpc.Dial(rpcAddr, opts...)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	return pb.NewByteGangsterClient(conn).Guacamole(context.Background(), &pb.ReqToken{Token: token})
}

// apiWsGuacamole websocket guacamole for RDP/VNC
// @Summary websocket guacamole for RDP/VNC
// @Description websocket guacamole for RDP/VNC
// @Tags Websocket
// @Accept  json
// @Produce  json
// @Param token query string true "TOKEN of gRPC to exchange asset's guacamole connection information"
// @Success 200 {object} responseObj
// @Router /api/ws/guacamole [get]
func (ap *app) apiWsGuacamole() func(c *gin.Context) {

	//0. 初始化 websocket 配置
	websocketReadBufferSize := guac.MaxGuacMessage
	websocketWriteBufferSize := guac.MaxGuacMessage * 2
	upgrade := websocket.Upgrader{
		ReadBufferSize:  websocketReadBufferSize,
		WriteBufferSize: websocketWriteBufferSize,
		CheckOrigin: func(r *http.Request) bool {
			//检查origin 限定websocket 被其他的域名访问
			return true
		},
	}
	return func(c *gin.Context) {

		token, isBreak := checkToken(c)
		if isBreak {
			return
		}

		//1. 解析参数, 因为 websocket 只能个通过浏览器url,request-header,cookie 传参数, 这里之接收 url-query 参数.
		//logrus.Println("1. 解析参数, 因为 websocket 只能个通过浏览器url,request-header,cookie 传参数, 这里之接收 url-query 参数.")
		arg, err := fetchGuacamoleArgFromRpc(token, ap.AddrRpc)
		if checkErr(c, err) {
			log.Println(err)
			return
		}

		//2. 设置为http-get websocket 升级
		//logrus.Println("2. 设置为http-get websocket 升级")
		protocol := c.GetHeader("Sec-Websocket-Protocol")
		ws, err := upgrade.Upgrade(c.Writer, c.Request, http.Header{
			"Sec-Websocket-Protocol": {protocol},
		})
		if err != nil {
			logrus.WithError(err).Error("升级ws失败")
			return
		}
		defer func() {
			if err = ws.Close(); err != nil {
				logrus.Traceln("Error closing websocket", err)
			}
		}()

		//3. 开始使用参数连接RDP远程桌面资产
		//logrus.Println("3. 开始使用参数连接RDP远程桌面资产, 对应guacamole protocol 文档的handshake章节")
		uid := ""
		//todo:: 让 guacamole支持更多协议
		pipeTunnel, err := guac.NewGuacamoleTunnel(ap.AddrGuacamole, arg.Protocol, arg.Host, arg.Port, arg.User, arg.Password, uid, int(arg.Width), int(arg.Height))
		if err != nil {
			logrus.Error("Failed to upgrade websocket", err)
			return
		}
		defer func() {
			if err = pipeTunnel.Close(); err != nil {
				logrus.Traceln("Error closing pipeTunnel", err)
			}
		}()
		//4. 开始处理 guacad-tunnel的io(reader,writer)
		//logrus.Println("4. 开始处理 guacad-tunnel的io(reader,writer)")
		//id := pipeTunnel.ConnectionID()

		ioCopy(ws, pipeTunnel)
		//logrus.Info("websocket session end")
	}
}

func ioCopy(ws *websocket.Conn, tnl *guac.SimpleTunnel) {

	writer := tnl.AcquireWriter()
	reader := tnl.AcquireReader()
	//if pipeTunnel.OnDisconnectWs != nil {
	//	defer pipeTunnel.OnDisconnectWs(id, ws, c.Request, pipeTunnel.TunnelPipe)
	//}
	defer tnl.ReleaseWriter()
	defer tnl.ReleaseReader()

	//使用 errgroup 来处理(管理) goroutine for-loop, 防止 for-goroutine zombie
	eg, _ := errgroup.WithContext(context.Background())

	eg.Go(func() error {
		buf := bytes.NewBuffer(make([]byte, 0, guac.MaxGuacMessage*2))

		for {
			ins, err := reader.ReadSome()
			if err != nil {
				return err
			}

			if bytes.HasPrefix(ins, guac.InternalOpcodeIns) {
				// messages starting with the InternalDataOpcode are never sent to the websocket
				continue
			}

			if _, err = buf.Write(ins); err != nil {
				return err
			}

			// if the buffer has more data in it or we've reached the max buffer size, send the data and reset
			if !reader.Available() || buf.Len() >= guac.MaxGuacMessage {
				if err = ws.WriteMessage(1, buf.Bytes()); err != nil {
					if err == websocket.ErrCloseSent {
						return fmt.Errorf("websocket:%v", err)
					}
					logrus.Traceln("Failed sending message to ws", err)
					return err
				}
				buf.Reset()
			}
		}

	})
	eg.Go(func() error {
		for {
			_, data, err := ws.ReadMessage()
			if err != nil {
				logrus.Traceln("Error reading message from ws", err)
				return err
			}
			if bytes.HasPrefix(data, guac.InternalOpcodeIns) {
				// messages starting with the InternalDataOpcode are never sent to guacd
				continue
			}
			if _, err = writer.Write(data); err != nil {
				logrus.Traceln("Failed writing to guacd", err)
				return err
			}
		}

	})
	if err := eg.Wait(); err != nil {
		logrus.WithError(err).Error("session-err")
	}

}
