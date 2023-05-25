package alipay

import (
	"context"
	"net/http"
	"testing"
)

func TestNotifyHandle(t *testing.T) {
	alipublicKey := `-----BEGIN PUBLIC KEY-----
	xxxxxx
	-----END PUBLIC KEY-----`
	a := NewAliPayReq("", "", "xxx", "123.1", "https://www.baidu.com")
	handler := NotifyHandle(alipublicKey, context.Background(), nil, a)
	err := http.ListenAndServe(":1234", handler)
	if err != nil {
		t.Error(err)
		return
	}
}
