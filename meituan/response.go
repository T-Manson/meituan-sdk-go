package meituan

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

// BaseResponse 基础响应
type BaseResponse interface {
	// 转Json字符串
	Json() string
	// 解析为struct
	Parse(jsonBytes []byte) error
}

// Response 单值响应
type Response struct {
	Data  string `json:"data"`
	Error *Error `json:"error,omitempty"`
}

func (resp Response) Json() string {
	jsonBytes, _ := json.Marshal(resp)
	return string(jsonBytes)
}

func (resp *Response) Parse(jsonBytes []byte) error {
	return json.Unmarshal(jsonBytes, resp)
}

// MapResponse 字典响应
type MapResponse struct {
	Data  map[string]interface{} `json:"data"`
	Error *Error                 `json:"error,omitempty"`
}

func (resp MapResponse) Json() string {
	jsonBytes, _ := json.Marshal(resp)
	return string(jsonBytes)
}

func (resp *MapResponse) Parse(jsonBytes []byte) error {
	return json.Unmarshal(jsonBytes, resp)
}

// ListMapResponse 字典集合响应
type ListMapResponse struct {
	Data  []map[string]interface{} `json:"data"`
	Error *Error                   `json:"error,omitempty"`
}

func (resp ListMapResponse) Json() string {
	jsonBytes, _ := json.Marshal(resp)
	return string(jsonBytes)
}

func (resp *ListMapResponse) Parse(jsonBytes []byte) error {
	return json.Unmarshal(jsonBytes, resp)
}

// Error 错误
type Error struct {
	Msg  string `json:"msg"`
	Code int    `json:"code"`
}

// SuccessResponse 成功响应
func SuccessResponse() string {
	response := &Response{Data: "ok"}
	return response.Json()
}

// ErrorResponse 异常响应
func ErrorResponse(code int, errorMsg string) string {
	response := &Response{Data: "ng", Error: &Error{Code: code, Msg: errorMsg}}
	return response.Json()
}

// ParseResponse 解析为指定的响应struct
//
// if has error, returns error else nil
func ParseResponse(resp *http.Response, outResp BaseResponse) error {
	var (
		result []byte
		err    error
	)

	if resp != nil {
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			if result, err = checkResponseBody(resp.Body); err != nil {
				logrus.Errorln("GetResponse checkResponseBody ", err.Error())
				_ = outResp.Parse(result)
				return err
			} else {
				logrus.Infoln("GetResponse Response ", string(result))
				return outResp.Parse(result)
			}
		} else {
			result, _ := ioutil.ReadAll(resp.Body)
			return fmt.Errorf("GetResponse code: %v body: %s", resp.StatusCode, string(result))
		}
	}

	logrus.Infoln("GetResponse Response is empty")
	return nil
}

// checkResponseBody 校验响应内容
//
// if error != nil, has error or response.data equals ng
func checkResponseBody(body io.ReadCloser) (result []byte, err error) {
	meiTuanResponse := &Response{}

	if result, err = ioutil.ReadAll(body); err != nil {
		return
	} else {
		if err = meiTuanResponse.Parse(result); err == nil {
			// 美团响应data为ng时，为处理失败
			if strings.ToLower(meiTuanResponse.Data) == "ng" {
				err = fmt.Errorf("checkResponseBody response.data equels ng. error:%+v", meiTuanResponse.Error)
				return
			}
		}
	}

	err = nil
	return
}
