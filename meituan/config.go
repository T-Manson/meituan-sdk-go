package meituan

// Config 美团配置
type Config struct {
	url            string
	appId          string
	consumerSecret string
}

func (cfg Config) Url() string {
	return cfg.url
}

func (cfg Config) AppId() string {
	return cfg.appId
}

func (cfg Config) ConsumerSecret() string {
	return cfg.consumerSecret
}

// NewConfig 构造配置
func NewConfig(url, appId, secret string) Config {
	return Config{
		url:            url,
		appId:          appId,
		consumerSecret: secret,
	}
}
