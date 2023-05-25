package alipay

import (
	"fmt"
	"testing"
)

func TestAliPayCommit(t *testing.T) {
	appID, privakey, publicKey := "xxx",
		`-----BEGIN RSA PRIVATE KEY-----
		-----END RSA PRIVATE KEY-----
		`,
		`-----BEGIN PUBLIC KEY-----
		MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCnxj/9qwVfgoUh/y2W89L6BkRA
		-----END PUBLIC KEY-----
		`
	a := NewAliPayReq("", "lalal", "xxxx", "12312.1", "www.baidu.com")
	uri, err := AliPayCommit(appID, privakey, publicKey, a)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(uri)
}
