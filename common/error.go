package common

/**
 * 消息错误提示
 */

const (
	WithoutError      = 0
	ParseError        = -32700 // 服务端接收到的json无法解析
	InvalidRequest    = -32600 // 发送无效的请求对象
	MethodNotFound    = -32601 // 该方法不存在或无效
	InvalidParams     = -32602 // 无效的方法参数
	InternalError     = -32603 // 内部调用错误
	ProcedureIsMethod = -32604 // 内部错误，请求未提供id字段
	CustomError       = -32000 // 服务端错误
)

var CodeMap = map[int]string{
	WithoutError:      "请求正常",
	ParseError:        "服务端接收到的json无法解析",
	InvalidRequest:    "发送无效的请求对象",
	MethodNotFound:    "该方法不存在或无效",
	InvalidParams:     "无效的方法参数",
	InternalError:     "内部调用错误",
	ProcedureIsMethod: "内部错误，请求未提供id字段",
	CustomError:       "服务端内部错误",
}
