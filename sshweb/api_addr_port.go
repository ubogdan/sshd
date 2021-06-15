package sshweb

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net"
	"time"
)

// apiTelnet test the sshd server is reachable to the addr
// @Summary test the sshd server is reachable to the addr
// @Description test the sshd server is reachable to the addr
// @Tags Util
// @Accept  json
// @Produce  json
// @Param addr query string true "the addr will be connected"
// @Success 200 {object} responseObj
// @Router /api/telnet [get]
func (ap *app) apiTelnet(c *gin.Context) {
	addr := c.Query("addr")
	conn, err := net.DialTimeout("tcp", addr, time.Millisecond*900)
	if checkErr(c, err) {
		return
	}
	conn.Close()
	returnData(c, fmt.Sprintf("%s is reachable", addr))
}
