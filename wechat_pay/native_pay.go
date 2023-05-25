package wechatpay

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
)

//API字典详情请查阅:https://pay.weixin.qq.com/wiki/doc/apiv3/apis/chapter3_4_1.shtml

//示例
/*
{
	"mchid": "1900006XXX",
	"out_trade_no": "native12177525012014070332333",
	"appid": "wxdace645e0bc2cXXX",
	"description": "Image形象店-深圳腾大-QQ公仔",
	"notify_url": "https://weixin.qq.com/",
	"amount": {
		"total": 1,
		"currency": "CNY"
	}
}
*/

// 服务端发起的支付请求
// 提供支付回调函数 -> 加密解密
// 1.调api生成预支付交易 -> 返回预支付交易链接(code_url)
// 2.异步通知商户支付结果，告知支付通知接收情况(apiv3密钥，解密得到json body)

const payTpye = "WechatPay"

type NativePay interface {
	GetNativeCodeUrl(appId, mchId, path string, option *Option) (*NativeRes, error)
}

var _ NativePay = &NativeReq{}

type Option struct {
	OnCallBack func(context.Context, ...interface{}) error
}

// Resource解密后的结构&Native下单req
type NativeReq struct {
	AppId          string       `json:"appid"`                      //应用ID
	MchId          string       `json:"mchid"`                      //直连商户号->由微信支付生成并下发
	Description    string       `json:"description,omitempty"`      //商品描述
	OutTradeNo     string       `json:"out_trade_no"`               //商户订单号->只能是数字、大小写字母_-*且在同一个商户号下唯一
	NotifyUrl      string       `json:"notify_url,omitempty"`       //通知地址
	Amount         NativeAmount `json:"amount"`                     //订单金额
	TimeExpire     string       `json:"time_expire,omitempty"`      //非必填 交易结束时间
	TransactionId  string       `json:"transaction_id,omitempty"`   //微信支付订单号
	TradeType      string       `json:"trade_type,omitempty"`       //交易类型 示例值：NATIVE
	TradeState     string       `json:"trade_state,omitempty"`      //交易状态 示例值：SUCCESS
	TradeStateDesc string       `json:"trade_state_desc,omitempty"` //交易状态描述
	BankType       string       `json:"bank_type,omitempty"`        //付款银行
	SuccessTime    string       `json:"success_time,omitempty"`     //支付完成时间
	Payer          Payer        `json:"payer,omitempty"`            //支付者
	IsHandle       bool         //是否已处理回调
	PayType        string       //支付类型 1.WechatPay 2.AliPay
	Debug          bool
}

// 支付者
type Payer struct {
	OpenId string `json:"openid"` //用户标识
}

// 订单金额
type NativeAmount struct {
	Total    float32 //是 总金额
	Currency string  //否 货币类型
}

type NativeRes struct {
	CodeUrl string `json:"code_url"` //是 二维码链接->并非固定值，按URL格式转换成二维码 示例值：weixin://wxpay/bizpayurl/up?pr=NwY5Mz9&groupid=00
}

func NewNativeReq(description, outTradeNo, notifyUrl string, amount NativeAmount) *NativeReq {
	return &NativeReq{
		Description: description,
		OutTradeNo:  outTradeNo,
		NotifyUrl:   notifyUrl,
		Amount:      amount,
	}
}

func Debug(debug bool, format string, a ...any) (n int, err error) {
	if debug {
		return fmt.Printf(format+"\n", a...)
	}
	return
}

/*
[GetNaticeCodeUrl]-> Native 预支付 POST https://api.mch.weixin.qq.com/v3/pay/transactions/native
appId:应用ID
mchId:商户号
path:本地文件中商户私钥的位置
*/
func (n *NativeReq) GetNativeCodeUrl(appId, mchId, path string, options *Option) (*NativeRes, error) {
	Debug(n.Debug, "WechatPrePay here")
	ctx := context.Background()
	// 1. 使用商户私钥等初始化 client，并使它具有自动定时获取微信支付平台证书的能力
	mchPrivateKey, err := utils.LoadPrivateKeyWithPath(path)
	if err != nil {
		fmt.Printf("Refund-> LoadPrivateKeyWithPath error(%v)", err)
		return nil, err
	}
	opts := []core.ClientOption{
		option.WithWechatPayAutoAuthCipher(mchID, mchCertificateSerialNumber, mchPrivateKey, mchAPIv3Key),
	}
	client, err := core.NewClient(ctx, opts...)
	if err != nil {
		fmt.Printf("Refund-> NewClient error(%v)", err)
		return nil, err
	}
	Debug(n.Debug, "Init client(%v) done", client)
	n.AppId = appId
	n.MchId = mchId
	n.PayType = payTpye //更新当前支付类型
	url := "https://api.mch.weixin.qq.com/v3/pay/transactions/native"
	result, err := client.Post(context.Background(), url, n)
	if err != nil {
		fmt.Printf("getNativeCodeUrl-> Post (%v) error(%v)", url, err)
		return nil, err
	}
	defer result.Response.Body.Close()
	body, err := ioutil.ReadAll(result.Response.Body)
	if err != nil {
		fmt.Printf("getNativeCodeUrl-> read response(%v) body error(%v)", result.Response, err)
		return nil, err
	}
	nativeRes := &NativeRes{}
	if err = json.Unmarshal(body, nativeRes); err != nil {
		fmt.Printf("getNativeCodeUrl-> Unmarshal body(%v) to nativeRes error(%v)", body, err)
		return nil, err
	}
	Debug(n.Debug, "Pre pay success, nativeRes: %v", nativeRes)
	return nativeRes, err
}

/*
[NativeCommit]->上层调用进行Native下单
appId:应用ID
mchId:商户号
path:本地文件中商户私钥的位置
*/
func NativeCommit(appId, mchId, path string, option *Option, nativeReq NativePay) (*NativeRes, error) {
	if nativeReq == nil {
		fmt.Printf("NativeCommit-> NativeReq can not be nil")
		return nil, errors.New("nativeCommit-> NativeReq can not be nil")
	}
	return nativeReq.GetNativeCodeUrl(appId, mchId, path, option)
}
