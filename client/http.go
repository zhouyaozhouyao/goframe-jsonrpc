package client

import (
	"bytes"
	"github.com/zhouyaozhouyao/goframe-jsonrpc/common"

	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

type Http struct {
	Ip          string
	Port        string
	RequestList []*common.SingleRequest
}

// NewHttpClient 实例化客户端对象
func NewHttpClient(ip string, port string) *Http {
	return &Http{
		Ip:          ip,
		Port:        port,
		RequestList: nil,
	}
}

func (p *Http) SetOptions(httpOptions interface{}) {

}

// BatchAppend 批量追加
func (p *Http) BatchAppend(method string, params interface{}, result interface{}, isNotify bool) *error {
	singleRequest := &common.SingleRequest{
		Method:   method,
		Params:   params,
		Result:   result,
		IsNotify: isNotify,
	}
	p.RequestList = append(p.RequestList, singleRequest)
	return singleRequest.Error
}

// BatchCall 批量调用
func (p *Http) BatchCall() error {
	var (
		err error
		br  []interface{}
	)
	for _, v := range p.RequestList {
		var req interface{}
		if v.IsNotify == true {
			req = common.Rs(nil, v.Method, v.Params)
		} else {
			req = common.Rs(strconv.FormatInt(time.Now().Unix(), 10), v.Method, v.Params)
		}
		br = append(br, req)
	}
	bReq := common.JsonBatchRs(br)
	err = p.handleFunc(bReq, p.RequestList)
	p.RequestList = make([]*common.SingleRequest, 0)
	return err
}

func (p *Http) Call(method string, params interface{}, result interface{}, isNotify bool) error {
	var (
		err error
		req []byte
	)

	if isNotify {
		req = common.JsonRs(nil, method, params)
	} else {
		req = common.JsonRs(strconv.FormatInt(time.Now().Unix(), 10), method, params)
	}
	err = p.handleFunc(req, result)
	return err
}

func (p *Http) handleFunc(b []byte, result interface{}) error {
	var url = fmt.Sprintf("http://%s:%s", p.Ip, p.Port)
	// 发送 POST 请求
	resp, err := http.Post(url, "application/json", bytes.NewReader(b)) //缓冲器 从一个[]byte切片，构造一个Buffer
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = common.GetResult(body, result)
	return err
}
