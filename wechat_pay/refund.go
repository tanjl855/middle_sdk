package wechatpay

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
)

//具体退款API详情及错误码请查阅:https://pay.weixin.qq.com/wiki/doc/apiv3/apis/chapter3_4_9.shtml

// 申请退款 https://api.mch.weixin.qq.com/v3/refund/domestic/refunds POST
type WechatRefund interface {
	Refund(path string) (*RefundResp, error)
}

var _ WechatRefund = &RefundReq{}

type RefundReq struct {
	OutTradeNo  string        `json:"out_trade_no"`          // 商户订单号
	OutRefundNo string        `json:"out_refund_no"`         // 商户退款单号
	Amount      *RefundAmount `json:"amount"`                // 金额信息
	Reason      string        `json:"reason"`                // [非必填]退款原因
	NotifyUrl   string        `json:"notify_url"`            // [非必填]退款结果回调url
	SuccessTime string        `json:"success_tim,omitempty"` // 支付完成时间
	Debug       bool
}

type RefundAmount struct {
	Refund   int    `json:"refund"`   //退款金额
	Total    int    `json:"total"`    //原订单金额
	Currency string `json:"currency"` //退款币种
}

type RefundResp struct {
	RefundId            string      `json:"refund_id"`             //微信支付退款单号
	OutRefundNo         string      `json:"out_refund_no"`         //商户退款单号
	TransactionId       string      `json:"transaction_id"`        //微信支付订单号
	OutTradeNo          string      `json:"out_trade_no"`          //商户订单号
	Channel             string      `json:"channel"`               //退款渠道
	UserReceivedAccount string      `json:"user_received_account"` //退款入账账户
	CreateTime          string      `json:"create_time"`           //退款创建时间
	Status              string      `json:"status"`                //退款状态
	Amount              *RespAmount `json:"amount"`                //金额信息
	FundsAccount        string      `json:"funds_account"`         //[非必填]资金账户
}

type RespAmount struct {
	Total            int    `json:"total"`             //订单金额
	Refund           int    `json:"refund"`            //退款金额
	PayerTotal       int    `json:"payer_total"`       //用户支付金额
	PayerRefund      int    `json:"payer_refund"`      //用户退款金额
	SettlementRefund int    `json:"settlement_refund"` //应结退款金额
	SettlementTotal  int    `json:"settlement_total"`  //应结订单金额
	DiscountRefund   int    `json:"discount_refund"`   //优惠退款金额
	Currency         string `json:"currency"`          //退款币种
}

func NewRefundReq(outTradeNo string, outRefundNo string, amount *RefundAmount) *RefundReq {
	return &RefundReq{
		OutTradeNo:  outTradeNo,
		OutRefundNo: outRefundNo,
		Amount:      amount,
	}
}

// path:本地文件中商户私钥的位置
func (refund *RefundReq) Refund(path string) (*RefundResp, error) {
	if !CheckDate(refund.SuccessTime) {
		return nil, errors.New("Refund->SuccessTime more than a year")
	}
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
	Debug(refund.Debug, "New client(%v) success", client)
	url := "https://api.mch.weixin.qq.com/v3/refund/domestic/refunds"
	result, err := client.Post(context.Background(), url, refund)
	if err != nil {
		fmt.Printf("Refund-> Post (%v) error(%v)", url, err)
		return nil, err
	}
	defer result.Response.Body.Close()
	Debug(refund.Debug, "Refund-> post by refund(%v),url(%v)", refund, url)

	body, err := ioutil.ReadAll(result.Response.Body)
	if err != nil {
		fmt.Printf("Refund-> read response(%v) body error(%v)", result.Response, err)
		return nil, err
	}
	fmt.Println(result.Response.Status)
	refundRes := &RefundResp{}
	if err = json.Unmarshal(body, refundRes); err != nil {
		fmt.Printf("Refund-> Unmarshal body(%v) to refundRes error(%v)", body, err)
		return nil, err
	}
	Debug(refund.Debug, "refund response(%v)", refundRes)
	return refundRes, err
}

/*
path:本地文件中商户私钥的位置
refundReq: 退款请求req
*/
func RefundCommit(path string, refundReq *RefundReq) (*RefundResp, error) {
	if refundReq == nil {
		fmt.Printf("RefundCommit-> refundReq can not be nil")
		return nil, errors.New("RefundCommit-> refundReq can not be nil")
	}
	return refundReq.Refund(path)
}

// check 订单完成时间是否超过一年,超过一年无法进行退款。
// 采用time.Time.Unix()换算,单位是s
func CheckDate(successTime string) bool {
	t, err := time.Parse(time.RFC3339, successTime)
	if err != nil {
		fmt.Printf("CheckDate-> Parse time error(%v)", err)
		return false
	}
	fmt.Println(t.Unix())
	fmt.Println(time.Now().Unix())
	return (time.Now().Unix() - t.Unix()) <= 3153600000
}
