package wechatpay

import (
	"fmt"
	"testing"
)

func TestCheckDate(t *testing.T) {
	nativeReq := &NativeReq{}
	nativeReq.SuccessTime = "2018-06-08T10:34:56+08:00"
	CheckDate(nativeReq.SuccessTime)
}

func TestRefundCommit(t *testing.T) {
	outTradeNo := "xxx"
	amount := &RefundAmount{}
	amount.Refund = 10000
	amount.Currency = "CNY"
	amount.Total = 10000
	refundReq := NewRefundReq(outTradeNo, "", amount)
	refundReq.SuccessTime = "2018-06-08T10:34:56+08:00"
	resp, err := RefundCommit(".././key.pem", refundReq)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(resp)
}
