package jsonrpc

import (
	"github.com/zhouyaozhouyao/goframe-jsonrpc/server"

	"errors"

	"golang.org/x/time/rate"
)

type ServerInterface interface {
	// SetBeforeFunc 方法执行前加载的函数
	SetBeforeFunc(func(id interface{}, method string, params interface{}) error)

	// SetAfterFunc 方法执行后加载的函数
	SetAfterFunc(func(id interface{}, method string, result interface{}) error)

	// SetOptions 在 Start 方法执行后执行添加可选参数
	SetOptions(interface{})

	// SetRateLimit 访问速率限制  使用 time/rate 限流器
	// rate.Limit 最大并发数量 10
	// int 每秒请求速率 20
	SetRateLimit(rate.Limit, int)

	// Start 启动入口
	Start()

	// Register jsonrpc 服务注册
	Register(s interface{})
}

func NewServer(protocol string, ip string, port string) (ServerInterface, error) {
	var err error
	// 根据 protocol 协议判断
	switch protocol {
	case "http":
		return server.NewHttpServer(ip, port), err
	case "tcp":
		return server.NewTcpServer(ip, port), err
	}
	return nil, errors.New("未找到匹配的协议")
}
