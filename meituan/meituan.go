package meituan

import (
	"fmt"
	"strings"
	"time"
)

const (
	// 常用美团接口路由常量
	MethodOrderBatchPullPhoneNumber = "order/batchPullPhoneNumber" // see: http://developer.waimai.meituan.com/home/docDetail/223
	MethodOrderCancel               = "order/cancel"               // see: http://developer.waimai.meituan.com/home/docDetail/114
	MethodOrderConfirm              = "order/confirm"              // see: http://developer.waimai.meituan.com/home/docDetail/111
	MethodOrderRefundReject         = "order/refund/reject"        // see: http://developer.waimai.meituan.com/home/docDetail/126
	MethodOrderRefundAgree          = "order/refund/agree"         // see: http://developer.waimai.meituan.com/home/docDetail/123
)

// 公共的配置变量
var commonConfig Config

// UseConfig 使用配置
func UseConfig(config Config) {
	commonConfig = config
}

// GetRequestUrl 获取请求美团的Url
func GetRequestUrl(method string) string {
	if strings.LastIndex(commonConfig.url, "/") < len(commonConfig.url)-1 {
		return fmt.Sprintf("%s/%s", commonConfig.url, method)
	}
	return fmt.Sprintf("%s%s", commonConfig.url, method)
}

// MakeTimestamp
func MakeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
