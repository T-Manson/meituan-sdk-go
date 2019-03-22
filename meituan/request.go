package meituan

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const HTTP_POST_CONTENT_TYPE = "application/x-www-form-urlencoded"

// MeiTuanRequest
type MeiTuanRequest struct {
	HttpMethod string
	RequestUrl string
	Timestamp  int64             // 调用接口时的时间戳，即当前时间戳（当前距离Epoch（1970年1月1日) 以秒计算的时间，即unix - timestamp），注意传输时间戳与当前北京时间前后相差不能超过10分钟
	AppId      string            // 美团分配给APP方的id
	Sig        string            // 输入参数计算后的签名结果
	Data       map[string]string // 应用级参数
}

// CallRemote
//
// outResponse Has three type can choose.
//
// type1: MeiTuanResponse.
//
// type2: MeiTuanMapResponse instead of func CallMapRemote().
//
// type3: MeiTuanListMapResponse  instead of func CallListMapRemote().
func (self *MeiTuanRequest) CallRemote(outResponse BaseMeiTuanResponse) (err error) {
	var response *http.Response

	if response, err = callMeiTuanApi(self); err != nil {
		return
	}

	err = GetMeiTuanResponse(response, outResponse)
	return
}

func (self *MeiTuanRequest) CallMapRemote() (*MeiTuanMapResponse, error) {
	meiTuanMapResponse := &MeiTuanMapResponse{}
	err := self.CallRemote(meiTuanMapResponse)
	return meiTuanMapResponse, err
}

func (self *MeiTuanRequest) CallListMapRemote() (*MeiTuanListMapResponse, error) {
	meiTuanListMapResponse := &MeiTuanListMapResponse{}
	err := self.CallRemote(meiTuanListMapResponse)
	return meiTuanListMapResponse, err
}

// SetData
func (self *MeiTuanRequest) SetData(data map[string]string) {
	self.Data = data
}

// ParseRequestParams
//
// if error != nil, has error
func (self *MeiTuanRequest) ParseRequestParams(requestBody string) error {
	var (
		timestamp, appId, sig string
		ok                    bool
		err                   error
	)

	if requestBody != "" {
		var values url.Values
		if values, err = url.ParseQuery(requestBody); err != nil {
			fmt.Println("[Error]ParseRequestParams ParseQuery ", err.Error())
			return err
		}

		var unescapeValuesStr1 string
		if unescapeValuesStr1, err = url.QueryUnescape(values.Encode()); err != nil {
			fmt.Println("[Error]ParseRequestParams QueryUnescape1 ", err.Error())
			return err
		}

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
		self.Data = data
	}

	if timestamp, ok = self.Data["timestamp"]; !ok {
		errMsg := "[Error]ParseRequestParams timestamp can not be empty"
		fmt.Println(errMsg)
		return fmt.Errorf(errMsg)
	}
	delete(self.Data, "timestamp")

	if appId, ok = self.Data["app_id"]; !ok {
		errMsg := "[Error]ParseRequestParams app_id can not be empty"
		fmt.Println(errMsg)
		return fmt.Errorf(errMsg)
	}
	delete(self.Data, "app_id")

	if sig, ok = self.Data["sig"]; !ok {
		errMsg := "[Error]ParseRequestParams sig can not be empty"
		fmt.Println(errMsg)
		return fmt.Errorf(errMsg)
	}
	delete(self.Data, "sig")

	if self.Timestamp, err = strconv.ParseInt(timestamp, 10, 64); err != nil {
		fmt.Println("[Error]ParseRequestParams ParseInt ", err.Error())
		return err
	}
	self.AppId = appId
	self.Sig = sig

	return nil
}

func (self *MeiTuanRequest) CheckPushSign() bool {
	sign, _, _ := makeSign(self.HttpMethod, self.RequestUrl,
		self.AppId, self.Timestamp, self.Data)
	if self.Sig == sign {
		return true
	}
	return false
}

func (self *MeiTuanRequest) GetDataValue(key string) interface{} {
	if self.Data != nil && len(self.Data) > 0 {
		return self.Data[key]
	}
	return nil
}

func (self *MeiTuanRequest) getFinalRequestUrl() (finalRequestUrl string, applicationParamStr string) {
	self.Timestamp = time.Now().UnixNano() / int64(time.Millisecond)
	self.AppId = commonConfig.appId

	var signValuesStr string
	self.Sig, signValuesStr, applicationParamStr = makeSign(self.HttpMethod, self.RequestUrl,
		self.AppId, self.Timestamp, self.Data)

	var finalRequestUrlValuesStr string
	switch self.HttpMethod {
	case http.MethodPost:
		finalRequestUrlValuesStr = fmt.Sprintf("%s?app_id=%s&timestamp=%v", self.RequestUrl, self.AppId, self.Timestamp)
	default:
		finalRequestUrlValuesStr = strings.Replace(signValuesStr, commonConfig.consumerSecret, "", -1)
	}
	finalRequestUrl = fmt.Sprintf("%s&sig=%s", finalRequestUrlValuesStr, self.Sig)
	return
}

// NewMeiTuanRequest returns *MeiTuanRequest
func NewMeiTuanRequest(httpMethod, requestUrl string) *MeiTuanRequest {
	return &MeiTuanRequest{
		HttpMethod: httpMethod,
		RequestUrl: requestUrl,
	}
}

// GetHttpUrlValues returns url.Values
func GetHttpUrlValues(dataMap map[string]string) (values url.Values) {
	var strBuilder strings.Builder
	for k, v := range dataMap {
		strBuilder.WriteString(fmt.Sprintf("%s=%v&", k, v))
	}
	if strBuilder.Len() > 0 {
		values, _ = url.ParseQuery(strBuilder.String())
	}
	return
}

// callMeiTuanApi
func callMeiTuanApi(self *MeiTuanRequest) (*http.Response, error) {
	var (
		response                             *http.Response
		finalRequestUrl, applicationParamStr string
		err                                  error
	)

	client := http.Client{}

	fmt.Println("[Info][MeiTuan]callMeiTuanApi requestUrl ", self.RequestUrl)
	finalRequestUrl, applicationParamStr = self.getFinalRequestUrl()
	fmt.Println("[Info][MeiTuan]callMeiTuanApi finalRequestUrl ", finalRequestUrl)

	switch self.HttpMethod {
	case http.MethodPost:
		fmt.Println("[Info][MeiTuan]callMeiTuanApi POST data: ", applicationParamStr)
		response, err = client.Post(finalRequestUrl, HTTP_POST_CONTENT_TYPE,
			strings.NewReader(applicationParamStr))
	default:
		response, err = client.Get(finalRequestUrl)
	}

	if err != nil {
		fmt.Println("[Error][MeiTuan]callMeiTuanApi ", self.RequestUrl, err.Error())
		return nil, err
	}

	return response, nil
}

// makeSign returns string, string, string
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
// appId: MeiTuan appId
//
// applicationParams: request data
//
// timestamp: request timestamp
//
// secret: MeiTuan secret
func makeSign(httpMethod, requestUrl,
	appId string,
	timestamp int64,
	requestData map[string]string) (sign string, signValuesStr string, applicationParamStr string) {
	if httpMethod == "" || requestUrl == "" || appId == "" || timestamp == 0 {
		return "", "", ""
	}

	signValuesStr, applicationParamStr = getSignValuesStr(httpMethod, requestUrl, appId, timestamp, requestData)
	fmt.Println("[Info][MeiTuan]makeSign sigValuesStr is: ", signValuesStr)
	md5Tool := md5.New()
	md5Tool.Write([]byte(signValuesStr))
	md5Bytes := md5Tool.Sum(nil)
	sign = hex.EncodeToString(md5Bytes)
	fmt.Println("[Info][MeiTuan]makeSign sign is: ", sign)
	return
}

// getSignValuesStr returns string, string
func getSignValuesStr(httpMethod, requestUrl,
	appId string, timestamp int64,
	requestData map[string]string) (sign string, applicationParamStr string) {
	values := GetHttpUrlValues(requestData)
	applicationParamStr = values.Encode()

	values.Add("timestamp", strconv.FormatInt(timestamp, 10))
	values.Add("app_id", appId)
	valuesStr, _ := url.QueryUnescape(values.Encode())
	sign = fmt.Sprintf("%s?%s%s", requestUrl, valuesStr, commonConfig.consumerSecret)
	return
}
