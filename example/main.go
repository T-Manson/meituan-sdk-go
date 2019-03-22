package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/T-Manson/meituan-sdk-go/meituan"
)

const (
	meiTuanUrl     = "https://waimaiopen.meituan.com/api/v1/"
	appId          = "AppId"
	consumerSecret = "Secret"
)

func main() {
	// 初始化
	meituan.UseConfig(meituan.NewConfig(meiTuanUrl, appId, consumerSecret))

	// 门店Code
	appPoiCode := "app_poi_code"

	// GET poi/mget 批量获取门店详细信息
	// see: http://developer.waimai.meituan.com/home/docDetail/9
	poiMget := "poi/mget"

	poiMgetRequest := meituan.NewRequest(http.MethodGet, meituan.GetRequestUrl(poiMget))
	poiMgetRequest.AddData("app_poi_codes", appPoiCode)

	// 获取结果的2种方式
	// 通过out参数返回
	poiMgetResponse := &meituan.ListMapResponse{}
	if err := poiMgetRequest.CallRemote(poiMgetResponse); err != nil {
		fmt.Println("[Error][MeiTuan]CallRemote ", poiMget, err.Error())
	} else {
		fmt.Println(poiMgetResponse.Json())
	}

	// 通过return结果返回
	if reps, err := poiMgetRequest.CallListMapRemote(); err != nil {
		fmt.Println("[Error][MeiTuan]CallRemote ", poiMget, err.Error())
	} else {
		fmt.Println(reps.Json())
	}

	// POST poi/online 门店设置为上线状态
	// see: http://developer.waimai.meituan.com/home/docDetail/21
	poiOnline := "poi/online"

	poiOnlineRequest := meituan.NewRequest(http.MethodGet, meituan.GetRequestUrl(poiOnline))
	poiOnlineRequest.AddData("app_poi_code", appPoiCode)

	poiOnlineResponse := &meituan.Response{}
	if err := poiOnlineRequest.CallRemote(poiOnlineResponse); err != nil {
		fmt.Println("[Error][MeiTuan]CallRemote ", poiOnline, err.Error())
	} else {
		fmt.Println(poiOnlineResponse.Json())
	}

	// 模拟一个美团请求过程
	// 美团推送请求： https://www.test.com/do/test 开发者在美团平台配置的推送地址
	// 美团推送请求方式： POST
	// 美团推送请求体： app_id=2788&order_id=11112222333344445&timestamp=1553222587
	requestUrl := "https://www.test.com/do/test"
	httpMethod := http.MethodPost
	values, _ := url.ParseQuery("app_id=2788&order_id=11112222333344445&timestamp=1553222587")

	formValuesTemp, _ := url.QueryUnescape(values.Encode())
	signValuesStr := requestUrl + "?" + formValuesTemp + consumerSecret
	md5Tool := md5.New()
	md5Tool.Write([]byte(signValuesStr))
	values.Add("sig", hex.EncodeToString(md5Tool.Sum(nil)))

	body := values.Encode()

	// 1. 将美团请求解析为struct
	meiTuanRequest := meituan.NewRequest(httpMethod, requestUrl)

	// 2. 解析美团请求体
	if err := meiTuanRequest.ParseRequestParams(body); err != nil {
		fmt.Println("[Error][MeiTuan]ParseRequestParams ", err.Error())
	}

	// 3. 验签
	checkResult := meiTuanRequest.CheckPushSign()
	fmt.Println("[Info][MeiTuan]CheckPushSign result: ", checkResult)

	// 4. Do something after CheckPushSign Success
	if checkResult {
		jsonBytes, _ := json.Marshal(meiTuanRequest)
		fmt.Println("[Info][MeiTuan]Request info: ", string(jsonBytes))
	}
}
