package alipay

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"time"
)

type AlipayRefund interface {
	AliPayRefund(appID, privateKey, aliPublicKey string) (*AliPayRefundRsp, error)
}

var _ AlipayRefund = &AliPayRefundReq{}

type AliPayRefundReq struct {
	AppAuthToken string // 可选 授权

	OutTradeNo string `json:"out_trade_no"` //商户订单号 与TradeNo二选一
	TradeNo    string `json:"trade_no"`     //支付宝交易号 与OutTradeNo二选一

	RefundAmount string `json:"refund_amount"`  //退款金额
	RefundReason string `json:"refund_reason"`  // 可选 退款的原因说明
	OutRequestNo string `json:"out_request_no"` // 必须 标识一次退款请求，同一笔交易多次退款需要保证唯一，如需部分退款，则此参数必传。
	Debug        bool
}

type AliPayRefundRsp struct {
	AlipayTradeRefundResponse AlipayResponse      `json:"alipay_trade_refund_response"`      //公共返回
	TradeNo                   string              `json:"trade_no"`                          // 支付宝交易号
	OutTradeNo                string              `json:"out_trade_no"`                      // 商户订单号
	BuyerLogonId              string              `json:"buyer_logon_id"`                    // 用户的登录id
	FundChange                string              `json:"fund_change"`                       // 本次退款是否发生了资金变化
	RefundFee                 string              `json:"refund_fee"`                        // 退款总金额
	RefundDetailItemList      []*RefundDetailItem `json:"refund_detail_item_list,omitempty"` // 退款使用的资金渠道
	StoreName                 string              `json:"store_name"`                        // 交易在支付时候的门店名称
	BuyerUserId               string              `json:"buyer_user_id"`                     // 买家在支付宝的用户id
	SendBackFee               string              `json:"send_back_fee"`                     // 退款总金额
}

type AlipayResponse struct {
	Code    string `json:"code"`     //网关返回码
	Msg     string `json:"msg"`      //网关返回码描述
	SubCode string `json:"sub_code"` //业务返回码
	SubMsg  string `json:"sub_msg"`  //业务返回码描述
	Sign    string `json:"sign"`     //签名
}

type RefundDetailItem struct {
	FundChannel string `json:"fund_channel"` // 交易使用的资金渠道，详见 支付渠道列表
	Amount      string `json:"amount"`       // 该支付工具类型所使用的金额
	RealAmount  string `json:"real_amount"`  // 渠道实际付款金额
	FundType    string `json:"fund_type"`    // 渠道所使用的资金类型
}

/*
outTradeNo和tradeNo二选一
outTradeNo:商户订单号
tradeNo:支付宝交易号
refundAmount:退款金额
refundReason:退款原因
outRequestNo:非必选，标识一次退款请求
*/
func NewAliPayRefundReq(outTradeNo, tradeNo, refundAmount, refundReason, outRequestNo string) *AliPayRefundReq {
	return &AliPayRefundReq{
		OutTradeNo:   outTradeNo,
		TradeNo:      tradeNo,
		RefundAmount: refundAmount,
		RefundReason: refundReason,
		OutRequestNo: outRequestNo,
	}
}

/*
appID: 支付宝分配给开发者的应用ID
privateKey: 开发者生成的私钥
aliPublicKey: 支付宝公钥
a: 支付宝支付请求struct
uri示例: https://openapi.alipay.com/gateway.do?timestamp=2013-01-01 08:08:08&method=alipay.trade.refund&app_id=19761&sign_type=RSA2&sign=ERITJKEIJKJHKKKKKKKHJEREEEEEEEEEEE&version=1.0&charset=GBK&biz_content=AlipayTradeRefundModel
*/
func (a *AliPayRefundReq) AliPayRefund(appID, privateKey, aliPublicKey string) (*AliPayRefundRsp, error) {
	client, err := GetAliPayClient(appID, privateKey, aliPublicKey)
	if err != nil {
		fmt.Printf("AliPayRefund-> get alipay client error(%v)", err)
		return nil, err
	}
	Debug(a.Debug, "AliPayRefund-> Get alipay client(%v) success", client)

	vals := url.Values{}
	vals.Add("charset", "utf-8")
	if a.OutTradeNo != "" {
		vals.Add("out_trade_no", a.OutTradeNo)
	} else {
		vals.Add("trade_no", a.TradeNo)
	}
	vals.Add("refund_amount", a.RefundAmount)
	vals.Add("refund_reason", a.RefundReason)
	if a.OutRequestNo != "" {
		vals.Add("out_request_no", a.OutRequestNo)
	}
	vals.Add("method", "alipay.trade.refund")
	vals.Add("app_id", appID)
	vals.Add("timestamp", time.Now().Format("2006-01-02 15:04:05"))
	vals.Add("charset", "utf-8")
	vals.Add("version", "1.0")
	var data = vals.Encode()
	data, _ = url.QueryUnescape(data)
	sign, _ := ShaSign(data, client.appPrivateKey)
	vals.Add("sign_type", "RSA2")
	vals.Add("sign", sign)
	Debug(a.Debug, "AliPayRefund-> add vals(%v) done", vals)

	uri := fmt.Sprintf("%v?%v", client.apiDomain, vals.Encode())
	Debug(a.Debug, "AliPayRefund-> create uri(%v) success", uri)

	resp, err := client.Client.Get(uri)
	if err != nil {
		fmt.Printf("AliPayRefund-> alipay postForm error(%v)", err)
		return nil, err
	}
	byteData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("AliPayRefund-> read resp body error(%v)", err)
		return nil, err
	}
	defer resp.Body.Close()
	refundRsp := &AliPayRefundRsp{}
	fmt.Println(string(byteData))
	if err = json.Unmarshal(byteData, refundRsp); err != nil {
		fmt.Printf("AliPayRefund-> Unmarshal byteData(%v) to refundRsp error(%v)", string(byteData), err)
		return nil, err
	}
	Debug(a.Debug, "AliPayRefund-> [Get] refundRsp(%v) by uri(%v) done", refundRsp, uri)
	return refundRsp, nil
}

// 上层调用
/*
appID: 支付宝分配给开发者的应用ID
privateKey: 开发者生成的私钥
aliPublicKey: 支付宝公钥
a: 支付宝支付请求struct
*/
func RefundByAliPay(appID, privateKey, aliPublicKey string, a *AliPayRefundReq) (*AliPayRefundRsp, error) {
	return a.AliPayRefund(appID, privateKey, aliPublicKey)
}
