package alipay

import (
	"fmt"
	"testing"
)

func TestRefundByAliPay(t *testing.T) {
	appID, privakey, publicKey := "xxxxx",
		`-----BEGIN RSA PRIVATE KEY-----
	-----END RSA PRIVATE KEY-----`,
		`-----BEGIN PUBLIC KEY-----
	xxx
	-----END PUBLIC KEY-----`
	a := NewAliPayRefundReq("xxx", "", "1231.123", "正常退款", "")
	resp, err := RefundByAliPay(appID, privakey, publicKey, a)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(resp)
}
