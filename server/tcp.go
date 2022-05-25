package server

import (
	"bytes"
	"context"
	"crm/jsonrpc/common"
	"fmt"
	"golang.org/x/time/rate"
	"log"
	"net"
	"sync"
)

type Tcp struct {
	Ip      string
	Port    string
	Server  common.Server
	Options TcpOptions
}

type TcpOptions struct {
	PackageEof       string
	PackageMaxLength int64
}

// NewTcpServer 建立 TcpServer 服务
func NewTcpServer(ip string, port string) *Tcp {
	options := TcpOptions{
		PackageEof:       "\r\n",
		PackageMaxLength: 1024 * 1024 * 2,
	}
	return &Tcp{
		Ip:   ip,
		Port: port,
		Server: common.Server{
			Sm:          sync.Map{},
			Hooks:       common.Hooks{},
			RateLimiter: nil,
		},
		Options: options,
	}
}

// Start 启动 tcp 服务
func (p *Tcp) Start() {
	var address = fmt.Sprintf("%s:%s", p.Ip, p.Port)
	tcpAddr, err := net.ResolveTCPAddr("tcp", address) // 解析 Tcp 服务
	if err != nil {
		common.Debug(err.Error())
	}

	listener, _ := net.ListenTCP("tcp", tcpAddr)
	log.Printf("Listening tcp://%s:%s", p.Ip, p.Port)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			common.Debug(err.Error())
			continue
		}
		go p.handleFunc(ctx, conn)
	}
}

// Register 注册服务
func (p *Tcp) Register(s interface{}) {
	_ = p.Server.Register(s)
}

func (p *Tcp) SetOptions(tcpOptions interface{}) {
	p.Options = tcpOptions.(TcpOptions)
}

// SetRateLimit 限流器
func (p *Tcp) SetRateLimit(r rate.Limit, b int) {
	p.Server.RateLimiter = rate.NewLimiter(r, b)
}

func (p *Tcp) SetBeforeFunc(beforeFunc func(id interface{}, method string, params interface{}) error) {
	p.Server.Hooks.BeforeFunc = beforeFunc
}

func (p *Tcp) SetAfterFunc(afterFunc func(id interface{}, method string, result interface{}) error) {
	p.Server.Hooks.AfterFunc = afterFunc
}

func (p *Tcp) handleFunc(ctx context.Context, conn net.Conn) {
	defer func(conn net.Conn) {
		_ = conn.Close()
	}(conn)

	select {
	// 监听超时或取消操作
	case <-ctx.Done():
		return
	default:

	}

	eofb := []byte(p.Options.PackageEof)
	eofl := len(eofb)
	for {
		var data []byte
		l := 0
		for {
			var buf = make([]byte, p.Options.PackageMaxLength)
			n, err := conn.Read(buf)
			if err != nil {
				if n == 0 {
					return
				}
				common.Debug(err.Error())
			}
			l += n
			data = append(data, buf[:n]...)
			// 检测两个字符串的长度和所包含的字节是否相同
			if bytes.Equal(data[l-eofl:], eofb) {
				break
			}
		}
		res := p.Server.Handler(data[:l-eofl])
		res = append(res, eofb...)
		_, _ = conn.Write(res)
	}
}
