# goframe-jsonrpc
基于 goframe 实现的 jsonrpc2.0可以和 hyperf 的 jsonrpc 无缝对接


```go
// 服务结构体
type IntRpc struct{}
// 参数
type Params struct {
   A int `json:"a"`
  B int `json:"b"`
}

type Result = g.Map // 返回结果

// 服务下的方法
func (i *IntRpc) Add(params *Params, result *Result) error {
   a := response.Success(g.Map{"value": params.A * params.B}, "请求成功")

   *result = interface{}(a).(Result)
   return nil
}

s, _ := jsonrpc.NewServer("http", "127.0.0.1", "8101") 建立连接
s.Register(new(IntRpc)) // 注册服务
s.Start() // 启动服务



// 客户端
c, _ := jsonrpc.NewClient("tcp", "127.0.0.1", "8101")
param := Params{2, 5}
result := new(Result) // 服务返回结果
err := c.Call("intRpc/add", &param, result, false) // 方法支持大驼峰，小驼峰，下划线
if err != nil {
return
}
g.Dump(*result)
```
