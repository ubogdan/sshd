package svr

import (
	"github.com/bytegang/pb"
	"google.golang.org/grpc"
)

func newRpcClient(addr string) (*grpc.ClientConn, pb.ByteGangsterClient, error) {
	var opts = []grpc.DialOption{grpc.WithInsecure()}
	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return nil, nil, err
	}
	return conn, pb.NewByteGangsterClient(conn), nil
}
