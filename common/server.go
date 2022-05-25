package common

import (
	"encoding/json"
	"fmt"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/util/gconv"
	"reflect"
	"strings"
	"sync"

	"golang.org/x/time/rate"
)

type Method struct {
	Name       string
	ParamsType reflect.Type
	ResultType reflect.Type
	Method     reflect.Method
}

// Service 服务实例
type Service struct {
	Name string
	V    reflect.Value
	T    reflect.Type
	Mm   map[string]*Method
}

// Server 服务
type Server struct {
	Sm          sync.Map      // 开启锁
	Hooks       Hooks         // 勾子函数
	RateLimiter *rate.Limiter // 限流器
}

type Hooks struct {
	BeforeFunc func(id interface{}, method string, params interface{}) error
	AfterFunc  func(id interface{}, method string, result interface{}) error
}

// Handler 处理参数与请求
func (svr *Server) Handler(b []byte) []byte {
	// 把参数解析为 json
	data, err := ParseRequestBody(b)
	if err != nil {
		return jsonE(nil, JsonRpc, ParseError)
	}
	var res interface{}
	if reflect.ValueOf(data).Kind() == reflect.Slice { // 检测参数类型是否为切片 slice 类型
		var resList []interface{}
		for _, v := range data.([]interface{}) {
			r := svr.SingleHandler(v.(map[string]interface{}))
			resList = append(resList, r)
		}
		res = resList
	} else if reflect.ValueOf(data).Kind() == reflect.Map { // 检测参数类型为 Map 类型
		r := svr.SingleHandler(data.(map[string]interface{}))

		res = r
	} else {
		return jsonE(nil, JsonRpc, InvalidRequest)
	}
	response, _ := json.Marshal(res)

	return response
}

func (svr *Server) Register(s interface{}) error {
	svc := new(Service)                              // 分配零值
	svc.V = reflect.ValueOf(s)                       // 获取值的对象
	svc.T = reflect.TypeOf(s)                        // 获取 interface 的具体类型
	svc.Name = reflect.Indirect(svc.V).Type().Name() // 返回 srv.V 指定的值 如果v是个nil指针，Indirect返回0值，如果v不是指针，Indirect返回v本身
	svc.Mm = RegisterMethods(svc.T)
	// 判断服务是否已经注册过
	if _, err := svr.Sm.LoadOrStore(svc.Name, svc); err {
		return gerror.New("当前服务已经注册过，请勿重新注册")
	}
	return nil
}

// RegisterMethods 注册方法
func RegisterMethods(s reflect.Type) map[string]*Method {
	mm := make(map[string]*Method, 0) // 初始化对象
	// s.NumMethod 获取方法数量
	for m := 0; m < s.NumMethod(); m++ {
		// s.Method(m) 循环遍历方法
		rm := s.Method(m)
		// 具体注册
		if mt := RegisterMethod(rm); mt != nil {
			mm[rm.Name] = mt
		}
	}
	return mm
}

func RegisterMethod(rm reflect.Method) *Method {
	var msg string
	rmt := rm.Type // 获取类型
	rmn := rm.Name // 获取名称
	// rm.NumIn 返回参数个数
	if rmt.NumIn() != 3 {
		msg = fmt.Sprintf("RegisterMethod：注册方法 %q 需要 %d 个参数", rmn, rmt.NumIn())
		Debug(msg)
		return nil
	}
	p := rmt.In(1) // 返回func类型的第i个参数的类型，如非函数或者i不在[0, NumIn())内将会panic
	// 判断第一个参数类型是否为指针类型
	if p.Kind() != reflect.Ptr {
		msg = fmt.Sprintf("RegisterMethod: 注册方法 %q的结果类型不是指针类型 %q", rmn, p)
		Debug(msg)
		return nil
	}

	r := rmt.In(2) // 返回func类型的第i个参数的类型，如非函数或者i不在[0, NumIn())内将会panic

	// Kind返回该接口的具体分类 不等于该指针
	if r.Kind() != reflect.Ptr {
		msg = fmt.Sprintf("RegisterMethod：注册方法 %q 的结果类型不是指针类型：%q", rmn, r)
		Debug(msg)
		return nil
	}

	// 检测函数的返回参数个数
	if rmt.NumOut() != 1 {
		msg = fmt.Sprintf("RegisterMethod: 注册方法 %q 返回参数个数不是一个 %q", rmn, rmt.NumOut())
		Debug(msg)
		return nil
	}

	// 返回func类型的第i个返回值的类型，如非函数或者i不在[0, NumOut())内将会panic
	ret := rmt.Out(0)
	// 判断返回参数的类型是否为nil
	if ret != reflect.TypeOf((*error)(nil)).Elem() {
		msg = fmt.Sprintf("RegisterMethod：注册方法 %q 的返回参数错误 %q", rmn, ret)
		Debug(msg)
		return nil
	}

	// 进行方法绑定
	m := &Method{
		Name:       rmn,
		ParamsType: p,
		ResultType: r,
		Method:     rm,
	}
	return m
}

func (svr *Server) SingleHandler(jsonMap map[string]interface{}) interface{} {

	id, jsonRpc, method, paramsData, errCode := ParseSingleRequestBody(jsonMap)
	if errCode != WithoutError {
		return E(id, jsonRpc, errCode)
	}

	if svr.RateLimiter != nil && !svr.RateLimiter.Allow() {
		return CE(id, jsonRpc, "请求次数过多，请稍候在试")
	}

	// 检测解析请求方法是否正确
	sName, mName, err := ParseRequestMethod(method)
	if err != nil {
		return E(id, JsonRpc, MethodNotFound)
	}
	// 读取 Map 中的方法
	s, ok := svr.Sm.Load(sName)
	if !ok {
		sName = lineToHump(sName) // 支持大驼峰、小驼峰、下划线
		s, ok = svr.Sm.Load(sName)
		if !ok {
			return E(id, jsonRpc,
				MethodNotFound)
		}
	}
	// 断言 s 对象是否为 Service 的实例
	m, ok := s.(*Service).Mm[mName]
	if !ok {
		return E(id, jsonRpc, MethodNotFound)
	}

	// 获取值类型并把所有属性的值分配零值
	params := reflect.New(m.ParamsType.Elem())
	pv := params.Interface() // 返回 interface 的 value 值
	// 转换
	err = gconv.Struct(paramsData, pv)
	if err != nil {
		return E(id, jsonRpc, InvalidParams)
	}
	// 获取返回参数的值并分配零值
	result := reflect.New(m.ResultType.Elem())

	// 检测是否开启了 before 操作  启动前的操作
	if svr.Hooks.BeforeFunc != nil {
		err = svr.Hooks.BeforeFunc(id, method, params.Elem().Interface())
		if err != nil {
			return CE(id, jsonRpc, err.Error())
		}
	}
	// Call 输入参数 in 并调用函数 v
	r := m.Method.Func.Call([]reflect.Value{s.(*Service).V, params, result})
	if i := r[0].Interface(); i != nil {
		Debug(i.(error))
		return E(id, jsonRpc, InternalError)
	}

	// 检测是否开启了 after 操作， 启动完成后的操作
	if svr.Hooks.AfterFunc != nil {
		err = svr.Hooks.AfterFunc(id, method, result.Elem().Interface())
		if err != nil {
			return CE(id, jsonRpc, err.Error())
		}
	}

	return S(id, jsonRpc, result.Elem().Interface())
}

func lineToHump(sName string) string {
	s := strings.Split(sName, "_")
	for k, v := range s {
		s[k] = Capitalize(v)
	}
	return strings.Join(s, "")
}

func Capitalize(str string) string {
	var upperStr string
	vv := []rune(str) // 处理中文 支持国际化语言
	for i := 0; i < len(vv); i++ {
		if i == 0 {
			if vv[i] >= 97 && vv[i] <= 122 {
				vv[i] -= 32
				upperStr += string(vv[i])
			} else {
				return str
			}
		} else {
			upperStr += string(vv[i])
		}
	}
	return upperStr
}
