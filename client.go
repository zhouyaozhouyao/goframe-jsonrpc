package jsonrpc

import (
	"github.com/zhouyaozhouyao/goframe-jsonrpc/client"
	"errors"
)

type ClientInterface interface {
	SetOptions(options interface{})                    // 调用 Call 或 BatchCall 前置操作
	Call(string, interface{}, interface{}, bool) error // 建立请求 支持 x/y 和 x.y
	BatchAppend(string, interface{}, interface{}, bool) *error
	BatchCall() error
}

func NewClient(protocol string, ip string, port string) (ClientInterface, error) {
	var err error
	switch protocol {
	case "http":
		return client.NewHttpClient(ip, port), err
	case "tcp":
		return client.NewTcpClient(ip, port)
	}
	return nil, errors.New("不支持当前协议")
}
