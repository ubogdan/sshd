package sshweb

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
)

func returnData(w *gin.Context, data interface{}) {
	w.JSON(200, responseObj{
		Code: 200,
		Msg:  "",
		Data: data,
	})

}

func returnErr(w *gin.Context, err interface{}) {

	w.JSON(200, responseObj{
		Code: 200,
		Msg:  fmt.Sprintf("%#v", err),
	})
}

func returnOk(w *gin.Context) {
	w.JSON(200, responseObj{
		Code: 200,
		Msg:  "ok",
	})
}

type responseObj struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func checkErr(w *gin.Context, err error) bool {
	flag := err != nil
	if flag {
		w.JSON(200, responseObj{
			Code: 207,
			Msg:  err.Error(),
		})
	}
	return flag
}

func returnFile(c *gin.Context, data []byte, ct, cd string) {
	buff := bytes.NewBuffer(data)
	c.DataFromReader(200, int64(buff.Len()), ct, buff, map[string]string{"Content-Disposition": cd})
}

func checkArgJson(c *gin.Context, ptr interface{}) bool {
	return checkErr(c, c.ShouldBind(ptr))
}

func checkToken(c *gin.Context) (token string, isBreak bool) {
	token = c.Query("token")
	if token == "" {
		returnErr(c, "token must not be empty")
		return "", true
	}

	return token, false
}
