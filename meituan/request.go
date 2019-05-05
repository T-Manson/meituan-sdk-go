package meituan

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// 美团POST请求正文类型
const HTTP_POST_CONTENT_TYPE = "application/x-www-form-urlencoded"

// Request 请求
type Request struct {
	HttpMethod string
	RequestUrl string
	Timestamp  int64             // 调用接口时的时间戳，即当前时间戳（当前距离Epoch（1970年1月1日) 以秒计算的时间，即unix - timestamp），注意传输时间戳与当前北京时间前后相差不能超过10分钟
	AppId      string            // 美团分配给APP方的id
	Sig        string            // 输入参数计算后的签名结果
	Data       map[string]string // 应用级参数
}

// CallRemote 远程调用，可返回单值类型、字典类型、字典集合类型响应结果
//
// outResp Has three type can choose.
//
// type1: Response.
//
// type2: MapResponse instead of func CallMapRemote().
//
// type3: ListMapResponse  instead of func CallListMapRemote().
func (req Request) CallRemote(outResp BaseResponse) (err error) {
	var resp *http.Response

	if resp, err = callApi(req); err != nil {
		return
	}

	err = ParseResponse(resp, outResp)
	return
}

func (req Request) CallMapRemote() (resp *MapResponse, err error) {
	resp = &MapResponse{}
	err = req.CallRemote(resp)
	return
}

func (req Request) CallListMapRemote() (resp *ListMapResponse, err error) {
	resp = &ListMapResponse{}
	err = req.CallRemote(resp)
	return
}

func (req Request) CheckPushSign() bool {
	if req.Sig == "" {
		return false
	}
	sign, _, _ := req.makeSign()
	return req.Sig == sign
}

func (req Request) GetDataValue(key string) string {
	if req.Data != nil {
		return req.Data[key]
	}
	return ""
}

func (req *Request) AddData(key string, value string) {
	if req.Data == nil {
		req.Data = make(map[string]string)
	}
	req.Data[key] = value
}

// ParseRequestParams 解析美团推送的请求体
//
// if error != nil, has error
func (req *Request) ParseRequestParams(reqBody string) error {
	var (
		timestamp, appId, sig string
		ok                    bool
		err                   error
	)

	if reqBody != "" {
		var values url.Values
		if values, err = url.ParseQuery(reqBody); err != nil {
			fmt.Println("[Error]ParseRequestParams ParseQuery ", err.Error())
			return err
		}

		// Why decode twice? see: http://developer.waimai.meituan.com/home/guide/6
		// 第一次解码
		var unescapeValuesStr1 string
		if unescapeValuesStr1, err = url.QueryUnescape(values.Encode()); err != nil {
			fmt.Println("[Error]ParseRequestParams QueryUnescape1 ", err.Error())
			return err
		}

		// 第二次解码
		var unescapeValuesStr2 string
		if unescapeValuesStr2, err = url.QueryUnescape(unescapeValuesStr1); err != nil {
			fmt.Println("[Error]ParseRequestParams QueryUnescape2 ", err.Error())
			return err
		}

		applicationParamValues, _ := url.ParseQuery(unescapeValuesStr2)

		data := make(map[string]string, len(applicationParamValues))
		for k, v := range applicationParamValues {
			data[k] = v[0]
		}
		req.Data = data
	}

	if timestamp, ok = req.Data["timestamp"]; !ok {
		errMsg := "[Error]ParseRequestParams timestamp can not be empty"
		fmt.Println(errMsg)
		return fmt.Errorf(errMsg)
	}
	delete(req.Data, "timestamp")

	if appId, ok = req.Data["app_id"]; !ok {
		errMsg := "[Error]ParseRequestParams app_id can not be empty"
		fmt.Println(errMsg)
		return fmt.Errorf(errMsg)
	}
	delete(req.Data, "app_id")

	if sig, ok = req.Data["sig"]; !ok {
		errMsg := "[Error]ParseRequestParams sig can not be empty"
		fmt.Println(errMsg)
		return fmt.Errorf(errMsg)
	}
	delete(req.Data, "sig")

	if req.Timestamp, err = strconv.ParseInt(timestamp, 10, 64); err != nil {
		fmt.Println("[Error]ParseRequestParams ParseInt ", err.Error())
		return err
	}
	req.AppId = appId
	req.Sig = sig

	return nil
}

func (req *Request) getFinalRequestUrl() (finalRequestUrl string, applicationParamStr string) {
	req.Timestamp = MakeTimestamp()
	req.AppId = commonConfig.appId

	var signValuesStr string
	req.Sig, signValuesStr, applicationParamStr = req.makeSign()

	var finalRequestUrlValuesStr string
	switch req.HttpMethod {
	case http.MethodPost:
		finalRequestUrlValuesStr = fmt.Sprintf("%s?app_id=%s&timestamp=%v", req.RequestUrl, req.AppId, req.Timestamp)
	default:
		finalRequestUrlValuesStr = strings.Replace(signValuesStr, commonConfig.consumerSecret, "", -1)
	}
	finalRequestUrl = fmt.Sprintf("%s&sig=%s", finalRequestUrlValuesStr, req.Sig)
	return
}

// makeSign 获取美团格式签名，返回：签名、签名使用的字符串、应用参数form格式字符串
//
// Sign Rule Exemple: ${proto}://${host}/${route}?app_id=${appId}&${applicationParams}&timestamp=${timestamp}${secret}
//
// proto: http/https
//
// host: domain
//
// route: internal operator
//
// url query: all key name must be asc sort
//
// appId:  appId
//
// applicationParams: request data
//
// timestamp: request timestamp
//
// secret:  secret
func (req *Request) makeSign() (sign string, signValuesStr string, applicationParamStr string) {
	if req.RequestUrl == "" || req.AppId == "" || req.Timestamp == 0 {
		return "", "", ""
	}

	signValuesStr, applicationParamStr = getSignValuesStr(req)
	fmt.Println("[Info][]makeSign sigValuesStr is: ", signValuesStr)
	md5Tool := md5.New()
	md5Tool.Write([]byte(signValuesStr))
	md5Bytes := md5Tool.Sum(nil)
	sign = hex.EncodeToString(md5Bytes)
	fmt.Println("[Info][]makeSign sign is: ", sign)
	return
}

// parseDataToHttpUrlValues 获取url query格式数据
func (req *Request) parseDataToHttpUrlValues() (values url.Values) {
	var strBuilder strings.Builder
	for k, v := range req.Data {
		strBuilder.WriteString(fmt.Sprintf("%s=%v&", k, v))
	}
	if strBuilder.Len() > 0 {
		values, _ = url.ParseQuery(strBuilder.String())
	}
	return
}

// NewRequest 构建请求
func NewRequest(httpMethod, requestUrl string) *Request {
	return &Request{
		HttpMethod: httpMethod,
		RequestUrl: requestUrl,
		Data:       make(map[string]string),
	}
}

// callApi 调用美团Api
func callApi(req Request) (*http.Response, error) {
	var (
		response                             *http.Response
		finalRequestUrl, applicationParamStr string
		err                                  error
	)

	client := http.Client{}

	fmt.Println("[Info][]callApi requestUrl ", req.RequestUrl)
	finalRequestUrl, applicationParamStr = req.getFinalRequestUrl()
	fmt.Println("[Info][]callApi finalRequestUrl ", finalRequestUrl)

	// 美团Api请求方式仅有Post、Get两种模式
	switch req.HttpMethod {
	case http.MethodPost:
		fmt.Println("[Info][]callApi POST data: ", applicationParamStr)
		response, err = client.Post(finalRequestUrl, HTTP_POST_CONTENT_TYPE,
			strings.NewReader(applicationParamStr))
	default:
		response, err = client.Get(finalRequestUrl)
	}

	if err != nil {
		fmt.Println("[Error][]callApi ", req.RequestUrl, err.Error())
		return nil, err
	}

	return response, nil
}

// getSignValuesStr 返回：签名使用的字符串、应用参数form格式字符串
func getSignValuesStr(req *Request) (signValuesStr string, applicationParamStr string) {
	values := req.parseDataToHttpUrlValues()
	applicationParamStr = values.Encode()

	values.Add("timestamp", strconv.FormatInt(req.Timestamp, 10))
	values.Add("app_id", req.AppId)
	valuesStr, _ := url.QueryUnescape(values.Encode())
	signValuesStr = fmt.Sprintf("%s?%s%s", req.RequestUrl, valuesStr, commonConfig.consumerSecret)
	return
}
