package kxsmsapi

import (
	"errors"
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
	dysmsapi "github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
)

// 1.阿里云短信服务
// SDK参考：https://help.aliyun.com/document_detail/112150.html?spm=a2c4g.215745.0.0.3ab66cf0XNSXe8
//部分错误码，具体请查阅https://help.aliyun.com/document_detail/101347.html?spm=a2c4g.101346.0.0.5ee37eb37ac1E2#section-91r-wtb-7z9
/*
Code								Message								   原因
OK									OK					   				   表示接口调用成功。
isv.SMS_SIGNATURE_SCENE_ILLEGAL		签名和模板类型不一致					模板和签名类型不一致，例如用验证码签名下发了通知短信或短信推广。
isv.EXTEND_CODE_ERROR				扩展码使用错误,							发送短信时，不同签名的短信使用了相同扩展码。
									相同的扩展码不可用于多个签名
isv.DOMESTIC_NUMBER_NOT_SUPPORTED	国际/港澳台消息模板不支持发送境内号码	 国际/港澳台消息模板仅支持发送国际、中国港澳台地区的号码。
isv.DENY_IP_RANGE					源IP地址所在的地区被禁用				被系统检测到源IP属于非中国内地地区。
isv.DAY_LIMIT_CONTROL				触发日发送限额							已经达到您在控制台设置的短信日发送量限额值。
isv.MONTH_LIMIT_CONTROL				触发月发送限额							已经达到您在控制台设置的短信月发送量限额值。
isv.SMS_CONTENT_ILLEGAL				短信内容包含禁止发送内容				 短信内容包含禁止发送内容。
isv.SMS_SIGN_ILLEGAL				签名禁止使用							签名禁止使用。
isp.RAM_PERMISSION_DENY				RAM权限不足								RAM权限不足。
isv.OUT_OF_SERVICE					业务停机								余额不足。业务停机时，套餐包中即使有短信额度也无法发送短信。
isv.PRODUCT_UN_SUBSCRIPT			未开通云通信产品的阿里云客户			该AccessKey所属的账号尚未开通云通信的服务，包括短信、语音、流量等服务。
isv.PRODUCT_UNSUBSCRIBE				产品未开通								该AccessKey所属的账号尚未开通当前接口的产品，例如仅开通了短信服务的用户调用语音服务接口时会产生此报错信息。
isv.ACCOUNT_NOT_EXISTS				账户不存在								使用了错误的账户名称或AccessKey。
isv.ACCOUNT_ABNORMAL				账户异常								账户异常。
isv.SMS_TEMPLATE_ILLEGAL			该账号下找不到对应模板					在您的账号下找不到对应模板，可能AccessKey账号和模板归属于不同账号，或使用了未审核通过的模板。
isv.SMS_SIGNATURE_ILLEGAL			该账号下找不到对应签名					在您的账号下找不到对应编号的签名，可能是AccessKey账号和签名归属于不同账号，或使用了未审核通过的签名。
																		   您传入的签名有空格、问号、错别字等导致乱码。
isv.INVALID_PARAMETERS				参数格式不正确							参数格式不正确。
isp.SYSTEM_ERROR					系统出现错误，请重新调用				系统出现错误。
isv.MOBILE_NUMBER_ILLEGAL			手机号码格式错误						手机号码格式错误。
isv.MOBILE_COUNT_OVER_LIMIT			手机号码数量超过限制，最多支持1000条	参数PhoneNumbers中指定的手机号码数量超出限制。
isv.TEMPLATE_MISSING_PARAMETERS		模板变量中存在未赋值变量				参数TemplateParam中，变量未全部赋值。
isv.BUSINESS_LIMIT_CONTROL			触发云通信流控限制						达到云通信短信发送频率上限。
isv.INVALID_JSON_PARAM				参数格式错误，请修改为字符串值			参数格式错误，不是合法的JSON格式，修改为字符串值。
SignatureNonceUsed					签名随机数已被使用						唯一随机数重复，SignatureNonce为唯一随机数，用于防止网络重放攻击。
*/

const (
	//初始化api默认信息
	product                  = "tanjlsmsapi" //名称
	version                  = "2024-05-26"  //版本号
	action                   = "SendSms"     //调用的api名称
	AliyunSmsRegionId        = "Bn-maoming"
	AliyunSmsAccessKeyId     = "xxxxxx"
	AliyunSmsAccessKeySecret = "xxx"
	ALIYUN                   = "Aliyun"
	Scheme                   = "https" // 默认使用https
)

// aliyun's request
type AliyunAdaptor struct {
	*requests.RpcRequest                  //请求参数*Method/Scheme...*
	Client               *dysmsapi.Client //Aliyun sdk's client
	Mock                 bool             //mock error
	CodeType             string           //需要mock的错误or成功返回Code值
	SmsUpExtendCode      string           `position:"Query" name:"SmsUpExtendCode"` //上行短信扩展码
	SignName             string           `position:"Query" name:"SignName"`        //是 签名
	PhoneNumbers         string           `position:"Query" name:"PhoneNumbers"`    //是 手机号
	OutId                string           `position:"Query" name:"OutId"`           //外部流水扩展字段
	TemplateCode         string           `position:"Query" name:"TemplateCode"`    //是 短信模板CODE
	TemplateParam        string           `position:"Query" name:"TemplateParam"`   //是 短信模板变量对应的实际值
	Debug                bool             //debug
}

// api's resposne
type AliyunSmsResponse struct {
	*responses.BaseResponse
	RequestId string `json:"RequestId" xml:"RequestId"` //请求ID
	BizId     string `json:"BizId" xml:"BizId"`         //发送回执ID
	Code      string `json:"Code" xml:"Code"`           //请求状态码
	Message   string `json:"Message" xml:"Message"`     //状态码描述
}

// init Aliyun Adaptor by schema, signName, templateCode and templateParam
/*
signName: 签名
templateCode: 短信模板CODE
templateParam: 短信模板变量对应的实际值
*/
func NewAliyunAdaptor(signName, templateCode, templateParam string) *AliyunAdaptor {
	aliyunAdaptor := &AliyunAdaptor{
		RpcRequest:    &requests.RpcRequest{},
		SignName:      signName,
		TemplateCode:  templateCode,
		TemplateParam: templateParam,
	}
	aliyunAdaptor.InitWithApiInfo(product, version, action, "", "")
	aliyunAdaptor.Method = requests.POST
	aliyunAdaptor.Scheme = Scheme
	return aliyunAdaptor
}

// Aliyun-> Send logic
func (adaptor *AliyunAdaptor) Send(phone string) (*SmsResponse, error) {
	Debug(adaptor.Debug, "Send message by Aliyun(%v)", adaptor)
	var err error
	adaptor.Client, err = dysmsapi.NewClientWithAccessKey(AliyunSmsRegionId,
		AliyunSmsAccessKeyId, AliyunSmsAccessKeySecret)
	if err != nil {
		return nil, err
	}
	Debug(adaptor.Debug, "Init client done!")
	//send api
	adaptor.PhoneNumbers = phone
	response := &AliyunSmsResponse{}
	response.BaseResponse = &responses.BaseResponse{}
	smsRes := &SmsResponse{}

	//mock error/OK here
	if adaptor.Mock && adaptor.CodeType != "" {
		smsRes.Code = adaptor.CodeType
		smsRes.Message = fmt.Sprintf("Mocking here,code:%v", smsRes.Code)
		if smsRes.Code != "OK" {
			return smsRes, errors.New(smsRes.Code)
		}
		return smsRes, nil
	}

	err = adaptor.Client.DoAction(adaptor, response)
	if err != nil {
		return nil, err
	}
	smsRes.Code = response.Code
	smsRes.Message = response.Message
	smsRes.RequestId = response.RequestId
	smsRes.BizId = response.BizId
	smsRes.SmsType = ALIYUN
	Debug(adaptor.Debug, "Send message success,smsRes: %v", smsRes)
	return smsRes, err
}
