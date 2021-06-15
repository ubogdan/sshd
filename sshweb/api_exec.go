package sshweb

import (
	"github.com/bytegang/pb"
	"github.com/bytegang/sshd/pkg/util"
	"github.com/gin-gonic/gin"
	"time"
)

func shellExec(in *pb.ReqShellExec) (res *pb.ResShellExec, err error) {
	startTs := time.Now().Unix()
	client, err := util.GetSshClient(in.Target, in.Proxy, time.Duration(in.TimeoutConnSec)*time.Second)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	out, err := util.SshRemoteRunCommandWithTimeout(client, in.ShellScript, time.Duration(in.TimeoutExecSec)*time.Second)

	result := pb.ResShellExec{
		Uuid:    in.Uuid,
		Output:  out,
		StartAt: uint64(startTs),
		DoneAt:  uint64(time.Now().Unix()),
		Code:    200,
		Msg:     "",
	}
	if err != nil {
		result.Output = err.Error()
		result.Ok = false
	} else {
		result.Output = out
		result.Ok = true
	}
	return &result, nil
}

// apiShellExec exec shell script by remote ssh
// @Summary test the sshd server is reachable to the addr
// @Description test the sshd server is reachable to the addr
// @Tags Util
// @Accept  json
// @Produce  json
// @Param arg body pb.ReqShellExec true "ssh connection information and script"
// @Success 200 {object} responseObj{data=pb.ResShellExec}
// @Router /api/ssh-exec [post]
func (ap *app) apiShellExec(c *gin.Context) {
	arg := new(pb.ReqShellExec)
	if checkArgJson(c, arg) {
		return
	}
	res, err := shellExec(arg)
	if checkErr(c, err) {
		return
	}
	returnData(c, res) //todo:: 统一json结构提, 可能有bug 需要优化
}
