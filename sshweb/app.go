package sshweb

import "github.com/gin-gonic/gin"

type app struct {
	router        *gin.Engine
	AddrRpc       string
	AddrGuacamole string
}

func NewApp(addrRpc, addrGuacamole string) *app {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	return &app{AddrRpc: addrRpc, router: r, AddrGuacamole: addrGuacamole}
}
