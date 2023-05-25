package kxsmsapi

import (
	"errors"
	"fmt"
)

type Adaptor interface {
	Send(phone string) (*SmsResponse, error)
}

var _ Adaptor = &AliyunAdaptor{}
var _ Adaptor = &TianYiyunAdaptor{}

type Config struct {
	OnInit    func() error       // 发送前调用
	OnRepeat  func() error       // [Aliyun]SignatureNoceUsed重复错误时调用
	OnError   func(string) error // 发送返回其他错误时调用 传入参数为smsRes.Code
	OnSuccess func(string) error // 发送成功时调用 传入参数为smsRes.SmsType（当前用的是阿里云or天翼云）
}

type SmsResponse struct {
	Code      string `json:"Code" xml:"Code"`                     //状态/天翼云要把desc->code
	Message   string `json:"Message" xml:"Message"`               //状态描述/desc
	TimeStamp string `json:"TimeStamp,omitempty" xml:"TimeStamp"` //格式yyyyMMddHHmmssSSS
	Sign      string `json:"Sign,omitempty" xml:"Sign"`           //签名(天翼云)
	BizId     string `json:"BizId,omitempty" xml:"BizId"`         //发送回执Id(阿里云)
	SmsType   string `json:"SmsType,omitempty" xml:"SmsType"`     //阿里云or天翼云
	RequestId string `json:"RequestId" xml:"RequestId"`           //消息标识/请求ID
}

func Debug(debug bool, format string, a ...any) (n int, err error) {
	if debug {
		return fmt.Printf(format+"\n", a...)
	}
	return
}

/*
phone: 手机号
config: hook
adaptor: 短信平台适配器(aliyun\tianyiyun)
*/
func SendSms(phone string, config *Config, adaptor ...Adaptor) (*SmsResponse, error) {
	// check args
	if len(adaptor) == 0 {
		fmt.Println("SendSms-> adaptor is nil")
		return nil, errors.New("adaptor is nil")
	}
	if phone == "" {
		fmt.Println("SendSms-> phone can not be empty")
		return nil, errors.New("phone is empty")
	}

	if config != nil && config.OnInit != nil {
		if err := config.OnInit(); err != nil {
			return nil, err
		}
	}

	var (
		err    error
		smsRes *SmsResponse
	)
	for i := 0; i < len(adaptor); i++ {
		smsRes, err = adaptor[i].Send(phone)
		if err != nil {
			return nil, err
		}
		if smsRes.Code == "SignatureNonceUsed" {
			//[阿里云] SignatureNonceUsed 重复了，重新发送
			if config != nil && config.OnRepeat != nil {
				if err := config.OnRepeat(); err != nil {
					return nil, err
				}
			}
			return SendSms(phone, config, adaptor...)
		}
		if smsRes.Code == "OK" {
			fmt.Println("SendSms-> send success!")
			if config != nil && config.OnSuccess != nil {
				if err := config.OnSuccess(smsRes.SmsType); err != nil {
					return nil, err
				}
			}
			return smsRes, nil
		}
		// Code != "OK" 时调用
		if config != nil && config.OnError != nil {
			if err := config.OnError(smsRes.Code); err != nil {
				return nil, err
			}
		}
	}
	if smsRes.Code != "OK" {
		return smsRes, errors.New(smsRes.Code)
	}
	return smsRes, nil
}
