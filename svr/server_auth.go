package svr

import (
	"context"
	"errors"
	pb "github.com/bytegang/pb"
	"github.com/bytegang/sshd/pkg/util"
	"golang.org/x/crypto/ssh"
	"log"
)

func getCtx() context.Context {
	return context.Background()
}

func (s *Server) authPublicKeys() func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
	return func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
		rpcConn, bastionClient, err := newRpcClient(s.Cfg.RpcAddr)
		if err != nil {
			return nil, err
		}
		defer rpcConn.Close()

		ins := pb.ReqAuthPublicKey{
			User: &pb.ReqSshUser{
				Account:       conn.User(),
				SessionId:     conn.SessionID(),
				ClientVersion: conn.ClientVersion(),
				ServerVersion: conn.ServerVersion(),
				RemoteAddr:    conn.RemoteAddr().String(),
				LocalAddr:     conn.LocalAddr().String(),
			},
			PublicKey: key.Marshal(),
		}
		user, err := bastionClient.AuthPk(getCtx(), &ins)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		return setPermission(user, util.Fingerprint(key), "public-key"), nil
	}
}

func (s *Server) authKeyboard() func(conn ssh.ConnMetadata, challenge ssh.KeyboardInteractiveChallenge) (*ssh.Permissions, error) {
	return func(conn ssh.ConnMetadata, challenge ssh.KeyboardInteractiveChallenge) (*ssh.Permissions, error) {
		rpcConn, bastionClient, err := newRpcClient(s.Cfg.RpcAddr)
		if err != nil {
			return nil, err
		}
		defer rpcConn.Close()
		ins := pb.ReqSshUser{
			Account:       conn.User(),
			SessionId:     conn.SessionID(),
			ClientVersion: conn.ClientVersion(),
			ServerVersion: conn.ServerVersion(),
			RemoteAddr:    conn.RemoteAddr().String(),
			LocalAddr:     conn.LocalAddr().String(),
		}

		kb, err := bastionClient.AuthKb(getCtx(), &ins)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		ans, err := challenge(kb.User, kb.Instruction, kb.Questions, kb.Echos)
		if err != nil {
			return nil, err
		}
		if len(ans) != len(kb.Answers) {
			return nil, errors.New("rpc auth-keyboard the count of returning questions and answers is not the same")
		}

		for i, an := range ans {
			if an != kb.Answers[i] {
				return nil, errors.New("answer is wrong")
			}
		}
		return setPermission(kb.ResUser, "", "keyboard"), nil
	}
}

func (s *Server) authPassword() func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {

	return func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
		//user := &pb.User{
		//	Id:      "1",
		//	Name:    "eric",
		//	Account: "zhou",
		//	Email:   "neochau@gmail.com",
		//	Phone:   "15527918920",
		//	Role:    1,
		//}
		//return setPermission(user, "", "password"), nil

		rpcConn, rpcClient, err := newRpcClient(s.Cfg.RpcAddr)
		if err != nil {
			return nil, err
		}
		defer rpcConn.Close()

		ins := pb.ReqAuthPassword{
			User: &pb.ReqSshUser{
				Account:       conn.User(),
				SessionId:     conn.SessionID(),
				ClientVersion: conn.ClientVersion(),
				ServerVersion: conn.ServerVersion(),
				RemoteAddr:    conn.RemoteAddr().String(),
				LocalAddr:     conn.LocalAddr().String(),
			},
			Password: password,
		}

		user, err := rpcClient.AuthPw(getCtx(), &ins)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		return setPermission(user, "", "password"), nil
	}
}
