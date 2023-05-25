package wechatpay

import (
	"fmt"
	"testing"
)

func TestNativeCommit(t *testing.T) {
	amount := NativeAmount{}
	amount.Total = 1231.11
	n := NewNativeReq("lalla", "123aba", "https://xxx.com", amount)
	appId, mchId, path := "", "", ".././xx.pem"
	res, err := NativeCommit(appId, mchId, path, nil, n)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(res)
}
