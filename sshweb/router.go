package sshweb

import (
	"github.com/bytegang/sshd/sshweb/frontendbuild"
	"net"
	_ "net/http/pprof"
)

// @title ByteGang Sshd
// @version 1.0
// @description This is a http server of bytegang sshd.
// @termsOfService http://swagger.io/terms/

// @contact.name EricZhou
// @contact.url https://mojotv.cn
// @contact.email neochau@gmail.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8022
// @BasePath /
// @schemes http

// @query.collection.format multi

//generate :  swag init -g sshweb/router.go --parseDependency

//help doc https://github.com/swaggo/swag/blob/master/README_zh-CN.md#api%E6%93%8D%E4%BD%9C

//Run run http server of bytegang sshd
func (ap *app) Run(lis net.Listener) error {
	frontendbuild.MwServeFrontendFiles(ap.router)
	// temporarily use relative path, run by `go run cmd/webshell/webshell_main.go` in project root path.
	// enter webshell by url like: http://127.0.0.1:8090/terminal?namespace=default&pod=nginx-65f9798fbf-jdrgl&container=nginx
	ap.router.GET("/api/ws/ssh", ap.wsTerminal())
	//router.HandleFunc("/api/ws/k8s-exec/{namespace}/{pod}/{container}", wsK8sXterm()).Methods("GET") //todo::k8s配置文件
	ap.router.GET("/api/ws/guacamole", ap.apiWsGuacamole()) //todo::支持 guacamole

	ap.router.GET("/api/sftp/ls", ap.apiSftpLs)
	ap.router.GET("/api/sftp/rm", ap.apiSftpRm)
	ap.router.POST("/api/sftp/upload", ap.apiSftpUpload)
	ap.router.GET("/api/sftp/mkdir", ap.apiSftpMkdir)
	ap.router.GET("/api/sftp/download/file", ap.apiSftpDownloadFile)
	ap.router.GET("/api/sftp/download/dir", ap.apiSftpDownloadDir)
	ap.router.GET("/api/telnet", ap.apiTelnet)
	ap.router.POST("/api/ssh-exec", ap.apiShellExec)
	return ap.router.RunListener(lis)
}

//http://127.0.0.1:8022/terminal?of=k8s-exec&t=k8s-exec&namespace=default&pod=hello-node-7567d9fdc9-7h6lc&container=echoserver&token=xxxxx

//done:: 支持ssh虚拟机

//http://127.0.0.1:8022/terminal?of=ssh&token=Ie2FcD_SKhTU7SpROGsuWTI3Oz3SLOvE4WBTPMHsDibPoQ9OtZwZcUvaf2vArk59TaDbeAwRY-3YKbNLR7rVCACa0lp7J6isY7Ypftq2zS8

//https://127.0.0.1:8022/#/xterm/
