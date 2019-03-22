package meituan

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

// BaseMeiTuanResponse
type BaseMeiTuanResponse interface {
	Json() string
	Parse(jsonBytes []byte) error
}

// MeiTuanResponse
type MeiTuanResponse struct {
	Data  string        `json:"data"`
	Error *MeiTuanError `json:"error,omitempty"`
}

// MeiTuanMapResponse
type MeiTuanMapResponse struct {
	Data  map[string]interface{} `json:"data"`
	Error *MeiTuanError          `json:"error,omitempty"`
}

// MeiTuanListMapResponse
type MeiTuanListMapResponse struct {
	Data  []map[string]interface{} `json:"data"`
	Error *MeiTuanError            `json:"error,omitempty"`
}

// MeiTuanError
type MeiTuanError struct {
	Msg  string `json:"msg"`
	Code int    `json:"code"`
}

func (self *MeiTuanResponse) Json() string {
	jsonBytes, _ := json.Marshal(self)
	return string(jsonBytes)
}

func (self *MeiTuanResponse) Parse(jsonBytes []byte) error {
	return json.Unmarshal(jsonBytes, self)
}

func (self *MeiTuanMapResponse) Json() string {
	jsonBytes, _ := json.Marshal(self)
	return string(jsonBytes)
}

func (self *MeiTuanMapResponse) Parse(jsonBytes []byte) error {
	return json.Unmarshal(jsonBytes, self)
}

func (self *MeiTuanListMapResponse) Json() string {
	jsonBytes, _ := json.Marshal(self)
	return string(jsonBytes)
}

func (self *MeiTuanListMapResponse) Parse(jsonBytes []byte) error {
	return json.Unmarshal(jsonBytes, self)
}

// GetSuccessResponse returns string
func GetSuccessResponse() string {
	response := &MeiTuanResponse{Data: "ok"}
	return response.Json()
}

// GetErrorResponse returns string
func GetErrorResponse(code int, errorMsg string) string {
	response := &MeiTuanResponse{Data: errorMsg, Error: &MeiTuanError{Code: code, Msg: errorMsg}}
	return response.Json()
}

// GetMeiTuanResponse
//
// if has error, returns error else nil
func GetMeiTuanResponse(resp *http.Response, outResponse BaseMeiTuanResponse) error {
	var (
		result []byte
		err    error
	)

	if resp != nil {
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			if result, err = checkMeiTuanResponseBody(resp.Body); err != nil {
				fmt.Println("[Error][MeiTuan]GetMeiTuanResponse checkMeiTuanResponseBody ", err.Error())
				return err
			} else {
				fmt.Println("[Info][MeiTuan]GetMeiTuanResponse Response ", string(result))
				return outResponse.Parse(result)
			}
		} else {
			result, _ := ioutil.ReadAll(resp.Body)
			return fmt.Errorf("[Error][MeiTuan]GetMeiTuanResponse code: %v body: %s", resp.StatusCode, string(result))
		}
	}

	fmt.Println("[Info][MeiTuan]GetMeiTuanResponse Response is empty")
	return nil
}

// checkMeiTuanResponseBody returns byte array
//
// if error != nil, has error or response.data equals ng
func checkMeiTuanResponseBody(body io.ReadCloser) (result []byte, err error) {
	meiTuanResponse := &MeiTuanResponse{}

	if result, err = ioutil.ReadAll(body); err != nil {
		return
	} else {
		if err = meiTuanResponse.Parse(result); err == nil {
			if strings.ToLower(meiTuanResponse.Data) == "ng" {
				result = nil
				err = fmt.Errorf("[Error][MeiTuan]checkMeiTuanResponseBody response.data equels ng. error:%+v", meiTuanResponse.Error)
				return
			}
		}
	}

	err = nil
	return
}
