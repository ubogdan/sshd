package svr

import "testing"

func TestRun(t *testing.T) {
	arg := Config{
		AppName:    "AnyOne",
		Bind:       ":2222",
		LogLvl:     "",
		Version:    "v1.0.1",
		LogDir:     ".",
		PrivateKey: defPK,
		RpcAddr:    "127.0.0.1:80999",
	}
	Run(&arg)
}
