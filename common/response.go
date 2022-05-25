package common

import (
	"encoding/json"
	"errors"
	"github.com/gogf/gf/v2/util/gconv"
	"reflect"
)

// SuccessResponse 成功响应
type SuccessResponse struct {
	Id      string      `json:"id"`
	JsonRpc string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
}

// SuccessNotifyResponse 异步请求成功响应
type SuccessNotifyResponse struct {
	JsonRpc string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
}

// Error 失败内容
type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Id      string `json:"id"`
	JsonRpc string `json:"jsonrpc"`
	Error   Error  `json:"error"`
}

// ErrorNotifyResponse 异常错误响应
type ErrorNotifyResponse struct {
	JsonRpc string `json:"jsonrpc"`
	Error   Error  `json:"error"`
}

// E 业务参数响应返回
func E(id interface{}, jsonRpc string, errCode int) interface{} {
	e := Error{
		Code:    errCode,
		Message: CodeMap[errCode],
		Data:    nil,
	}

	var res interface{}
	if id != nil {
		res = ErrorResponse{Id: id.(string), JsonRpc: jsonRpc, Error: e}
	} else {
		res = ErrorNotifyResponse{JsonRpc: jsonRpc, Error: e}
	}
	return res
}

func CE(id interface{}, jsonRpc string, errMessage string) interface{} {
	e := Error{
		Code:    CustomError,
		Message: errMessage,
		Data:    nil,
	}
	var res interface{}
	if id != nil {
		res = ErrorResponse{id.(string), jsonRpc, e}
	} else {
		res = ErrorNotifyResponse{jsonRpc, e}
	}
	return res
}

func S(id interface{}, jsonRpc string, result interface{}) interface{} {
	var res interface{}
	if id != nil {
		res = SuccessResponse{id.(string), jsonRpc, result}
	} else {
		res = SuccessNotifyResponse{jsonRpc, result}
	}
	return res
}

// JsonE 标准格式响应返回
func jsonE(id interface{}, jsonRpc string, errCode int) []byte {
	e, _ := json.Marshal(E(id, jsonRpc, errCode))
	return e
}

// GetResult 获取接口返回信息
func GetResult(b []byte, result interface{}) error {
	var (
		err      error
		jsonData interface{}
	)
	err = json.Unmarshal(b, &jsonData)
	if err != nil {
		Debug(err)
	}

	// 判断返回值类型是否为 Map
	if reflect.ValueOf(jsonData).Kind() == reflect.Map {
		if err = GetSingleResponse(jsonData.(map[string]interface{}), result); err != nil {
			return err
		}
	} else if reflect.ValueOf(jsonData).Kind() == reflect.Slice {
		for k, v := range jsonData.([]interface{}) {
			err = GetSingleResponse(v.(map[string]interface{}), result.([]*SingleRequest)[k].Result)
			if err != nil {
				*(result.([]*SingleRequest)[k].Error) = err
			}
		}
	}
	return nil
}

// GetSingleResponse 获取单一请求的响应
func GetSingleResponse(jsonMap map[string]interface{}, result interface{}) error {
	var err error
	emData, ok := jsonMap["error"]
	if ok {
		resErr := new(Error) // 分配一个零值的 Error
		err = GetStruct(emData, resErr)
		Debug(resErr.Message)
		return errors.New(resErr.Message)
	}
	// 处理返回结果值
	if err = gconv.Struct(jsonMap["result"], result); err != nil {
		Debug(err)
		return err
	}
	return err
}
