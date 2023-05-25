package kxsmsapi

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestSendSms(t *testing.T) {
	code := "1234"
	aliyunAdaptor := NewAliyunAdaptor("xxx", "SMS_Send", fmt.Sprintf(`{"code":"%v"}`, code))
	config := &Config{
		OnInit: func() error {
			fmt.Println("Init")
			return nil
		},
		OnSuccess: func(la string) error {
			fmt.Println(la)
			return nil
		},
	}
	aliyunAdaptor.Debug = true

	var err error
	//
	_, err = SendSms("18200771880", config, aliyunAdaptor)
	if err != nil {
		t.Error(err)
		return
	}
	//mock error
	//
	aliyunAdaptor.Mock = true
	aliyunAdaptor.CodeType = "isv.SMS_SIGNATURE_ILLEGAL"
	_, err = SendSms("18200771880", config, aliyunAdaptor)
	if err == nil {
		t.Error("isv.SMS_SIGNATURE_ILLEGAL error but no return err")
	}
	//
	aliyunAdaptor.CodeType = "isv.MOBILE_NUMBER_ILLEGAL"
	_, err = SendSms("18200771880", config, aliyunAdaptor)
	if err == nil {
		t.Error("isv.MOBILE_NUMBER_ILLEGAL error but no return err")
		return
	}
	//
	aliyunAdaptor.CodeType = "isv.SMS_TEMPLATE_ILLEGAL"
	_, err = SendSms("18200771880", config, aliyunAdaptor)
	if err == nil {
		t.Error("isv.SMS_TEMPLATE_ILLEGAL error but no return err")
		return
	}
	//
	aliyunAdaptor.CodeType = "isv.TEMPLATE_PARAMS_ILLEGAL"
	_, err = SendSms("18200771880", config, aliyunAdaptor)
	if err == nil {
		t.Error("isv.TEMPLATE_PARAMS_ILLEGAL error but no return err")
		return
	}
	//mock成功->
	aliyunAdaptor.CodeType = "OK"
	_, err = SendSms("18200771880", config, aliyunAdaptor)
	if err != nil {
		t.Error(err)
		return
	}
	aliyunAdaptor.Mock = false

}

func TestTianYiyun(t *testing.T) {
	// 天翼云
	code := "6666"
	tianYiyunAdaptor := NewTianYiyunAdaptor("111.11.121.151:1234", "xxx", "xxxx", "xxxx", "【tanjl】您的短信验证码：%v，该验证码5分钟内有效，请勿泄露于他人！", code)
	tianYiyunAdaptor.Debug = true
	res, err := SendSms("18200771880", nil, tianYiyunAdaptor)
	if err != nil {
		t.Error(err)
		return
	}
	byteRes, err := json.Marshal(res)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Printf("response:%v", string(byteRes))
}
