package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gogf/gf/v2/util/gconv"
	"reflect"
	"strings"
)

const (
	JsonRpc = "2.0" // 标准协议版本号
)

// RequiredFields 请求体的标准参数
var RequiredFields = map[string]string{
	"id":      "id",
	"jsonrpc": "jsonrpc",
	"method":  "method",
	"params":  "params",
}

// Request 请求参数列表
type Request struct {
	Id      string      `json:"id"`
	JsonRpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

// NotifyRequest 异常请求参数列表
type NotifyRequest struct {
	JsonRpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

// SingleRequest 客户端参数结构体
type SingleRequest struct {
	Method   string      // 请求方式
	Params   interface{} // 请求参数
	Result   interface{} // 响应结果
	Error    *error      // 错误
	IsNotify bool        // 是否异步
}

// ParseRequestBody 解析字符串为 json
func ParseRequestBody(b []byte) (interface{}, error) {
	var jsonData interface{}
	err := json.Unmarshal(b, &jsonData)
	if err != nil {
		Debug(err)
	}
	return jsonData, err
}

// ParseSingleRequestBody 解析请求参数
func ParseSingleRequestBody(jsonMap map[string]interface{}) (id interface{}, jsonRpc string, method string, params interface{}, errCode int) {
	jsonMap = FilterRequestBody(jsonMap)
	if _, ok := jsonMap["id"]; ok != true {
		st := NotifyRequest{}
		if err := gconv.Struct(jsonMap, &st); err != nil {
			// 参数转换异常
			Debug(err)
			errCode = InvalidRequest
		}
		return nil, st.JsonRpc, st.Method, st.Params, errCode
	} else {
		st := Request{}
		if err := gconv.Struct(jsonMap, &st); err != nil {
			errCode = InvalidRequest
		}
		return st.Id, st.JsonRpc, st.Method, st.Params, errCode
	}
}

// FilterRequestBody 过滤请求参数
func FilterRequestBody(jsonMap map[string]interface{}) map[string]interface{} {
	for k, _ := range jsonMap {
		if _, ok := RequiredFields[k]; !ok {
			delete(jsonMap, k)
		}
	}
	return jsonMap
}

// ParseRequestMethod 解析请求方法
func ParseRequestMethod(method string) (sName string, mName string, err error) {
	var (
		m  string
		sp int
	)
	first := method[0:1]
	if first == "." || first == "/" {
		method = method[1:]
	}
	if strings.Count(method, ".") != 1 && strings.Count(method, "/") != 1 {
		m = fmt.Sprintf("rpc：方法 %s 请求格式错误，需要为 x.y 或 x/y ", method)
		Debug(m)
		return sName, mName, errors.New(m)
	}

	if strings.Count(method, ".") == 1 {
		sp = strings.LastIndex(method, ".")
		if sp < 0 {
			m = fmt.Sprintf("rpc: 方法 %s 请求格式错误; 需要为 x.y or x/y", method)
			return sName, mName, errors.New(m)
		}
		sName = method[:sp]
		mName = method[sp+1:]
	} else if strings.Count(method, "/") == 1 {
		sp = strings.LastIndex(method, "/")
		if sp < 0 {
			m = fmt.Sprintf("rpc: 方法 %s 请求格式错误; 需要为 x.y or x/y", method)
			return sName, mName, errors.New(m)
		}
		sName = method[:sp]
		mName = method[sp+1:]
	}
	return sName, lineToHump(mName), err
}

// Rs 参数组装
func Rs(id interface{}, method string, params interface{}) interface{} {
	var req interface{}
	if id != nil {
		req = Request{Id: id.(string), JsonRpc: JsonRpc, Method: method, Params: params}
	} else {
		req = NotifyRequest{JsonRpc: JsonRpc, Method: method, Params: params}
	}
	return req
}

func JsonBatchRs(data []interface{}) []byte {
	e, _ := json.Marshal(data)
	return e
}

func GetStruct(d interface{}, s interface{}) error {
	var (
		m string
		t reflect.Type
	)

	// 检测类型是否为Ptr
	if reflect.TypeOf(s).Kind() != reflect.Ptr {
		m = fmt.Sprintf("无效的类型元素 %s，需要类型为指针", reflect.TypeOf(s))
		Debug(m)
		return errors.New(m)
	}
	t = reflect.TypeOf(s).Elem() // 获取值类型的数据  TypeOf 返回接口值类型 Elem 获取指向类型的数据
	var jsonMap = make(map[string]interface{}, 0)
	// reflect.TypeOf(d).Kind() 返回具体类型
	switch reflect.TypeOf(d).Kind() {
	case reflect.Map:
		if t.NumField() != len(d.(map[string]interface{})) {
			m = fmt.Sprintf("json：参数个数不匹配")
			Debug(m)
			return errors.New(m)
		}
		for k := 0; k < t.NumField(); k++ {
			lk := strings.ToLower(t.Field(k).Name) // 转换字符为小写
			if _, ok := d.(map[string]interface{})[lk]; ok != true {
				m = fmt.Sprintf("json: 找不到该字段 %s", lk)
				Debug(m)
				return errors.New(m)
			}
		}
		jsonMap = d.(map[string]interface{})
		break
	case reflect.Slice:
		// 字段数量不匹配
		if t.NumField() != reflect.ValueOf(d).Len() {
			m = fmt.Sprintf("参数数量不匹配 [%s] [%s]", t.NumField(), reflect.ValueOf(d).Len())
			Debug(m)
			return errors.New(m)
		}

		for k := 0; k < t.NumField(); k++ {
			jsonMap[t.Field(k).Name] = reflect.ValueOf(d).Index(k).Interface()
		}
		break
	default:
		break
	}
	if err := gconv.Struct(jsonMap, s); err != nil {
		Debug(err)
		return err
	}
	return nil
}

func JsonRs(id interface{}, method string, params interface{}) []byte {
	e, _ := json.Marshal(Rs(id, method, params))
	return e
}
