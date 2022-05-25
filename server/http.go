package server

import (
	"crm/jsonrpc/common"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

type Http struct {
	Ip      string
	Port    string
	Server  common.Server
	Options HttpOptions
}

type HttpOptions struct {
}

// NewHttpServer 启动入口
func NewHttpServer(ip string, port string) *Http {
	options := HttpOptions{}
	return &Http{
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

// Start 启动 http 连接
func (p *Http) Start() {
	// 自定义多路由分发服务
	mux := http.NewServeMux()
	// 注册根路由
	mux.HandleFunc("/", p.handleFunc)
	// 启动服务
	var url = fmt.Sprintf("%s:%s", p.Ip, p.Port)
	log.Printf("Listening http://%s:%s", p.Ip, p.Port)
	_ = http.ListenAndServe(url, mux)
}

func (p *Http) SetBeforeFunc(beforeFunc func(id interface{}, method string, params interface{}) error) {
	p.Server.Hooks.BeforeFunc = beforeFunc
}

func (p *Http) SetAfterFunc(afterFunc func(id interface{}, method string, result interface{}) error) {
	p.Server.Hooks.AfterFunc = afterFunc
}

func (p *Http) SetOptions(httpOptions interface{}) {
	p.Options = httpOptions.(HttpOptions)
}

func (p *Http) SetRateLimit(r rate.Limit, b int) {
	// 限流器
	p.Server.RateLimiter = rate.NewLimiter(r, b)
}

func (p *Http) Register(s interface{}) {
	_ = p.Server.Register(s)
}

// handleFunc 注册路由
func (p *Http) handleFunc(w http.ResponseWriter, r *http.Request) {
	var (
		err  error
		data []byte
	)
	// 添加请求头类型
	w.Header().Set("Content-Type", "application/json")
	// 读取文件或网络请求 ioutil.ReadAll
	if data, err = ioutil.ReadAll(r.Body); err != nil {
		// 响应状态码 500 服务器异常
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if r.Method != http.MethodPost {
		// 响应状态码 405 请求方法不存在
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	resp := p.Server.Handler(data)
	// 返回结果
	_, _ = w.Write(resp)
}
