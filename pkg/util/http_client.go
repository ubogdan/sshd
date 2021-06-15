package util

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

//httpClient 跳过Tls检验的golang标准库客户端
var httpClient = func() http.Client {
	//https://stackoverflow.com/questions/12122159/how-to-do-a-https-request-with-bad-certificate
	//tr := &http.Transport{
	//	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	//}
	return http.Client{Timeout: time.Second * 20}
}()

//RequestJson 发送json参数
func RequestJson(url, method string, reqBody interface{}, headersKv map[string]string) (*http.Response, error) {

	bs, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(bs))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	for hk, hv := range headersKv {
		req.Header.Set(hk, hv)
	}
	return httpClient.Do(req)
}

//RequestJsonString 发送json-string参数
func RequestJsonString(url, method string, reqBody string, headersKv map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, bytes.NewBufferString(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for hk, hv := range headersKv {
		req.Header.Set(hk, hv)
	}
	return httpClient.Do(req)
}

//ResponseUnmarshal 解析http response中的json
func ResponseUnmarshal(resp *http.Response, vPtr interface{}) (bs []byte, err error) {
	bs, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	err = json.Unmarshal(bs, vPtr)
	return bs, err
}

func SendRequest(url, method string, reqBody interface{}, headersKv map[string]string, resPtr interface{}) (bs []byte, err error) {
	response, err := RequestJson(url, method, reqBody, headersKv)
	if err != nil {
		return nil, err
	}
	return ResponseUnmarshal(response, resPtr)
}
