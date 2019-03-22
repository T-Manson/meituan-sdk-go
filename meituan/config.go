package meituan

// Config 美团配置
type Config struct {
	url            string
	appId          string
	consumerSecret string
}

func (self *Config) GetUrl() string {
	return self.url
}

func (self *Config) GetAppId() string {
	return self.appId
}

func (self *Config) GetConsumerSecret() string {
	return self.consumerSecret
}

// NewConfig 构造配置
func NewConfig(url, appId, secret string) Config {
	return Config{
		url:            url,
		appId:          appId,
		consumerSecret: secret,
	}
}
