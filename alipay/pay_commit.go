package alipay

import (
	"fmt"
	"net/url"
	"time"
)

// 支付宝支付文档详情:https://opendocs.alipay.com/open/270/01didh?pathHash=a6ccbe9a&ref=api#%E6%8E%A5%E5%8F%A3%E8%B0%83%E7%94%A8%E9%85%8D%E7%BD%AE
// 支付宝网关: https://openapi.alipay.com/gateway.do

type AliPay interface {
	PayCommit(appID, privateKey, aliPublicKey string) (string, error)
}

var _ AliPay = &AliPayReq{}

type AliTrade struct {
	NotifyURL    string // 支付宝服务器主动通知商户服务器里指定的页面http/https路径。
	ReturnURL    string // 支付成功后跳转界面url
	AppAuthToken string // 可选

	OutTradeNo  string `json:"out_trade_no"` //商户订单号 64个字符以内，仅支持字母、数字、下划线且需保证在商户端不重复。
	TradeNo     string `json:"trade_no"`     //支付宝交易号
	TotalAmount string `json:"total_amount"` //订单总金额 单位为元，精确到小数点后两位，取值范围[0.01,100000000]
	Subject     string `json:"subject"`      //订单标题
	ProductCode string `json:"product_code"` //销售产品码 目前电脑支付场景下仅支持FAST_INSTANT_TRADE_PAY

	GoodsDetail     []*GoodsDetail `json:"goods_detail,omitempty"`      // 可选 订单包含的商品列表信息，Json格式，详见商品明细说明
	TimeExpire      string         `json:"time_expire,omitempty"`       // 可选 订单绝对超时时间 格式为yyyy-MM-dd HH:mm:ss。超时时间范围：1m~15d。
	BusinessParams  string         `json:"business_params,omitempty"`   // 可选 商户传入业务信息，具体值要和支付宝约定，应用于安全，营销等参数直传场景，格式为json格式
	PromoParams     string         `json:"promo_params,omitempty"`      // 可选 优惠参数 注：仅与支付宝协商后可用
	InvoiceInfo     string         `json:"invoice_info,omitempty"`      // 可选 开票信息
	ExtUserInfo     *ExtUserInfo   `json:"ext_user_info,omitempty"`     // 可选 外部指定买家
	MerchantOrderNo string         `json:"merchant_order_no,omitempty"` // 可选 商户的原始订单号
	StoreId         string         `json:"store_id,omitempty"`          // 可选 商户门店编号。该参数用于请求参数中以区分各门店，非必传项。
}

// 可选 外部指定买家
type ExtUserInfo struct {
	Name          string `json:"name"`            //  可选 指定买家姓名。 注： need_check_info=T时该参数才有效
	Mobile        string `json:"mobile"`          //  可选 指定买家手机号。 注：该参数暂不校验
	CertType      string `json:"cert_type"`       //  可选 指定买家证件类型。 枚举值：IDENTITY_CARD：身份证；PASSPORT：护照；OFFICER_CARD：军官证；SOLDIER_CARD：士兵证；HOKOU：户口本。如有其它类型需要支持，请与蚂蚁金服工作人员联系。注： need_check_info=T时该参数才有效，支付宝会比较买家在支付宝留存的证件类型与该参数传入的值是否匹配。
	CertNo        string `json:"cert_no"`         //  可选 买家证件号。 注：need_check_info=T时该参数才有效，支付宝会比较买家在支付宝留存的证件号码与该参数传入的值是否匹配。
	MinAge        string `json:"min_age"`         //  可选 允许的最小买家年龄。 买家年龄必须大于等于所传数值注：1. need_check_info=T时该参数才有效  2. min_age为整数，必须大于等于0
	NeedCheckInfo string `json:"need_check_info"` //  可选 是否强制校验买家信息； 需要强制校验传：T;不需要强制校验传：F或者不传；当传T时，支付宝会校验支付买家的信息与接口上传递的cert_type、cert_no、name或age是否匹配，只有接口传递了信息才会进行对应项的校验；只要有任何一项信息校验不匹配交易都会失败。如果传递了need_check_info，但是没有传任何校验项，则不进行任何校验。默认为不校验。
	IdentityHash  string `json:"identity_hash"`   //  可选 买家加密身份信息。当指定了此参数且指定need_check_info=T时，支付宝会对买家身份进行校验，校验逻辑为买家姓名、买家证件号拼接后的字符串，以sha256算法utf-8编码计算hash，若与传入的值不匹配则会拦截本次支付。注意：如果同时指定了用户明文身份信息（name，cert_type，cert_no中任意一个），则忽略identity_hash以明文参数校验。
}

// 可选 订单包含的商品列表信息
type GoodsDetail struct {
	GoodsId        string  `json:"goods_id"`
	AliPayGoodsId  string  `json:"alipay_goods_id,omitempty"`
	GoodsName      string  `json:"goods_name"`
	Quantity       int     `json:"quantity"`
	Price          float64 `json:"price"`
	GoodsCategory  string  `json:"goods_category,omitempty"`
	CategoriesTree string  `json:"categories_tree,omitempty"`
	Body           string  `json:"body,omitempty"`
	ShowURL        string  `json:"show_url,omitempty"`
}

// AliPay请求结构
type AliPayReq struct {
	AliTrade
	QrPayMode   string `json:"qr_pay_mode"`  //可选 PC扫码支付的方式 支持前置模式和跳转模式
	QrCodeWidth int    `json:"qrcode_width"` //可选 商户自定义二维码宽度 qr_pay_mode=4时该参数生效
	PayStatus   string //支付状态 1.交易创建，等待买家付款 2.未付款交易超时关闭，或支付完成后全额退款 3.交易支付成功 4.交易结束，不可退款
	PayType     string //支付产品类型 1.WechatPay 2.AliPay
	Debug       bool
}

// AliPay返回
type AliPayResp struct {
	Code            string `json:"code"`              //网关返回码
	Msg             string `json:"msg"`               //网关返回码描述
	SubCode         string `json:"sub_code"`          //可选 业务返回码
	SubMsg          string `json:"sub_msg"`           //可选 业务返回码描述
	TradeNo         string `json:"trade_no"`          // 支付宝交易号
	OutTradeNo      string `json:"out_trade_no"`      // 商家订单号
	SellerId        string `json:"seller_id"`         // 收款支付宝账号对应的支付宝唯一用户号，以2088开头的纯16位数字
	TotalAmount     string `json:"total_amount"`      // 交易的订单金额
	MerchantOrderNo string `json:"merchant_order_no"` // 商户原始订单号，最大长度限制32位
}

const (
	aliPay      = "AliPay"
	ProductCode = "FAST_INSTANT_TRADE_PAY" // 目前电脑支付场景下仅支持FAST_INSTANT_TRADE_PAY
)

func NewAliPayReq(notifyUrl, subject, outTradeNo, totalAmount, returnURL string) *AliPayReq {
	return &AliPayReq{
		AliTrade: AliTrade{
			Subject:     subject,
			OutTradeNo:  outTradeNo,
			TotalAmount: totalAmount,
			NotifyURL:   notifyUrl,
			ReturnURL:   returnURL,
		},
	}
}

/*
appID: 支付宝分配给开发者的应用ID
privateKey: 开发者生成的私钥
aliPublicKey: 支付宝公钥
return示例：https://openapi.alipay.com/gateway.do?timestamp=2013-01-01 08:08:08&method=alipay.trade.page.pay&app_id=24610&sign_type=RSA2&sign=ERITJKEIJKJHKKKKKKKHJEREEEEEEEEEEE&version=1.0&charset=GBK&biz_content=AlipayTradePageCreateandpayModel
*/
func (a *AliPayReq) PayCommit(appID, privateKey, aliPublicKey string) (string, error) {
	a.PayType = aliPay
	client, err := GetAliPayClient(appID, privateKey, aliPublicKey)
	if err != nil {
		fmt.Printf("PayCommit-> Get alipay error(%v)", err)
		return "", err
	}
	Debug(a.Debug, "PayCommit-> Get alipay client(%v) success", client)

	var vals = &url.Values{}
	vals.Add("method", "alipay.trade.page.pay")
	vals.Add("app_id", appID)
	vals.Add("timestamp", time.Now().Format("2006-01-02 15:04:05"))
	vals.Add("charset", "utf-8")
	vals.Add("version", "1.0")
	vals.Add("notify_url", a.NotifyURL)
	if len(a.ReturnURL) > 0 {
		vals.Add("return_url", a.ReturnURL)
	}
	vals.Add("out_trade_no", a.OutTradeNo)
	vals.Add("total_amount", a.TotalAmount)
	vals.Add("subject", a.Subject)
	vals.Add("product_code", a.ProductCode)

	var data = vals.Encode()
	data, _ = url.QueryUnescape(data)
	sign, _ := ShaSign(data, client.appPrivateKey)
	vals.Add("sign_type", "RSA2")
	vals.Add("sign", sign)
	Debug(a.Debug, "PayCommit-> add vals(%v) done", vals)

	uri := fmt.Sprintf("%v?%v", client.apiDomain, vals.Encode())
	Debug(a.Debug, "PayCommit-> create uri(%v) success", uri)
	return uri, nil
}

// 上层调用
/*
appID: 支付宝分配给开发者的应用ID
privateKey: 开发者生成的私钥
aliPublicKey: 支付宝公钥
isProduction: false表示支付宝沙箱环境，true表示生产环境
a: 支付宝支付请求struct
返回支付界面url
*/
func AliPayCommit(appID, privateKey, aliPublicKey string, a *AliPayReq) (string, error) {
	return a.PayCommit(appID, privateKey, aliPublicKey)
}
