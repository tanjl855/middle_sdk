package alipay

import (
	"crypto/rsa"
	"fmt"
	"net/http"
	"sync"
)

type Client struct {
	mux       sync.Mutex
	appId     string
	apiDomain string
	Client    *http.Client

	appPrivateKey    *rsa.PrivateKey
	aliPublicCertSN  string
	aliPublicKeyList map[string]*rsa.PublicKey
}

func Debug(debug bool, format string, a ...any) (n int, err error) {
	if debug {
		return fmt.Printf(format+"\n", a...)
	}
	return
}

func WithHttpClient(client *http.Client) OptionFunc {
	return func(c *Client) {
		c.Client = client
	}
}

type OptionFunc func(c *Client)

func NewAlipayClient(appId, privateKey string, opts ...OptionFunc) (client *Client, err error) {
	priKey, err := ParsePKCS1PrivateKey(FormatPKCS1PrivateKey(privateKey))
	if err != nil {
		return nil, err
	}
	client = &Client{}
	client.appId = appId

	client.apiDomain = "https://openapi.alipay.com/gateway.do"
	client.Client = http.DefaultClient
	client.appPrivateKey = priKey
	client.aliPublicKeyList = make(map[string]*rsa.PublicKey)

	for _, opt := range opts {
		opt(client)
	}

	return client, nil
}

// 生成Client
/*
appID:支付宝分配给开发者的应用ID
privateKey: 开发者生成的私钥
aliPublicKey: 支付宝公钥
*/
func GetAliPayClient(appID, privateKey, aliPublicKey string) (*Client, error) {
	var client, err = NewAlipayClient(appID, privateKey)
	if err != nil {
		fmt.Printf("GetAliPayClient-> New alipay's client failed, error(%v)", err)
		return nil, err
	}
	err = client.LoadAliPayPublicKey(aliPublicKey)
	if err != nil {
		fmt.Printf("GetAliPayClient-> load alipay publicKey(%v) error(%v)", aliPublicKey, err)
		return nil, err
	}
	return client, nil
}
