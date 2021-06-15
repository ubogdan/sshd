package svr

import (
	"encoding/json"
	"github.com/bytegang/pb"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

const permissionUserIns = "authedUserJson"

//getPermissionUser 获取登录的用户信息
func getPermissionUser(conn *ssh.ServerConn) (*pb.User, error) {
	userString := conn.Permissions.Extensions[permissionUserIns]
	user := new(pb.User)
	err := json.Unmarshal([]byte(userString), user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

//setPermission 设置登录的用户信息 给认证之后的session使用
func setPermission(user *pb.User, fp, form string) *ssh.Permissions {
	marshal, err := json.Marshal(user)
	if err != nil {
		logrus.Error(err)
	}
	return &ssh.Permissions{Extensions: map[string]string{permissionUserIns: string(marshal), "from": form, "fingerprint": fp}}
}
