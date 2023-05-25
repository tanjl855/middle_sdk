package wechatpay

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/wechatpay-apiv3/wechatpay-go/core/auth/verifiers"
	"github.com/wechatpay-apiv3/wechatpay-go/core/downloader"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
)

//API字典详情请查阅https://pay.weixin.qq.com/wiki/doc/apiv3/apis/chapter3_4_5.shtml
//微信会对发送给商户的通知进行签名，将签名值放在通知的HTTP头Wechatpay-Signature

/*
resource 解密 -> NativeReq  示例如下：
{
    "transaction_id":"1217752501201407033233368018",
    "amount":{
        "payer_total":100,
        "total":100,
        "currency":"CNY",
        "payer_currency":"CNY"
    },
    "mchid":"1230000109",
    "trade_state":"SUCCESS",
    "bank_type":"CMC",
    "success_time":"2018-06-08T10:34:56+08:00",
    "payer":{
        "openid":"oUpF8uMuAJO_M2pxb1Q9zNjWeS6o"
    },
    "out_trade_no":"1217752501201407033233368018",
    "appid":"wxd678efh567hg6787",
    "trade_state_desc":"支付成功",
    "trade_type":"MICROPAY",
}
*/

// 支付通知req -> 对resource解密 -> NativeReq
type NotifyReq struct {
	ID           string   `json:"id,omitempty"`  // 通知ID
	CreateTime   string   `json:"create_time"`   // 通知创建的时间 eg: 2015-05-20T13:29:35+08:00
	EventType    string   `json:"event_type"`    // 通知类型	支付成功通知的类型为TRANSACTION.SUCCESS
	ResourceType string   `json:"resource_type"` // 通知数据类型	 eg: encrypt-resource
	Resource     Resource `json:"resource"`      // 通知数据
	Summary      string   `json:"summary"`       // 回调摘要	示例值：支付成功
}

// 通知数据
type Resource struct {
	Algorithm      string `json:"algorithm"`       //加密算法类型 示例值：AEAD_AES_256_GCM
	Ciphertext     string `json:"ciphertext"`      //数据密文 示例值：sadsadsadsad 使用key、nonce和associated_data解密得到json资源对象
	OriginalType   string `json:"original_type"`   //原始类型 示例值：transaction
	Nonce          string `json:"nonce"`           //随机串 示例值：fdasflkja484w
	AssociatedData string `json:"associated_data"` //非必填 附加数据
	Plaintext      string // Ciphertext 解密后内容
}

// 支付通知res
type NotifyRes struct {
	Code string `json:"code"`    // 返回状态码 错误码，SUCCESS为清算机构接收成功，其他错误码为失败.示例值：FAIL
	Msg  string `json:"message"` // 返回信息 示例值：失败
}

const (
	mchID                      string = "190000****"                               // 商户号
	mchCertificateSerialNumber string = "3775B6A45ACD588826D15E583A95F5DD********" // 商户证书序列号
	mchAPIv3Key                string = "2ab9****************************"         // 商户APIv3密钥
)

/*
[NotifyHandle]
处理微信回调通知,把body解密->传入的nativeReq
ctx: 上下文信息
path: 示例 "/path/to/merchant/apiclient_key.pem"
options: 提供钩子函数
notifyReq: 支付通知请求,resource未解密
nativeReq: resource解密后结构
*/
func NotifyHandle(ctx context.Context, path string, options *Option, myNotifyReq *NotifyReq, nativeReq *NativeReq) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//检查业务数据
		if nativeReq.IsHandle {
			w.WriteHeader(http.StatusOK)
			return
		}
		//回调通知的验签和解密
		mchPrivateKey, err := utils.LoadPrivateKeyWithPath(path) //从本地加载商户私钥
		if err != nil {
			fmt.Printf("load merchant private key error")
			w.WriteHeader(http.StatusInternalServerError)
			errRes := &NotifyRes{
				Code: "FAIL",
				Msg:  err.Error(),
			}
			errByte, err := json.Marshal(errRes)
			if err != nil {
				fmt.Printf("Marshal errRes(%v) error(%v)", errRes, err)
				return
			}
			w.Write(errByte)
			return
		}
		// 1. 使用 `RegisterDownloaderWithPrivateKey` 注册下载器
		err = downloader.MgrInstance().RegisterDownloaderWithPrivateKey(ctx, mchPrivateKey, mchCertificateSerialNumber, mchID, mchAPIv3Key)
		if err != nil {
			fmt.Printf("RegisterDownloaderWithPrivateKey error(%v)", err)
			w.WriteHeader(http.StatusInternalServerError)
			errRes := &NotifyRes{
				Code: "FAIL",
				Msg:  err.Error(),
			}
			errByte, err := json.Marshal(errRes)
			if err != nil {
				fmt.Printf("Marshal errRes(%v) error(%v)", errRes, err)
				return
			}
			w.Write(errByte)
			return
		}
		// 2. 获取商户号对应的微信支付平台证书访问器
		certificateVisitor := downloader.MgrInstance().GetCertificateVisitor(mchID)
		// 3. 使用证书访问器初始化 `notify.Handler`
		handler := notify.NewNotifyHandler(mchAPIv3Key, verifiers.NewSHA256WithRSAVerifier(certificateVisitor))

		//获取notifyReq和解密Resourc -> nativeReq
		var notifyReq *notify.Request
		notifyReq, err = handler.ParseNotifyRequest(ctx, r, nativeReq)
		if err != nil {
			fmt.Printf("ParseNotifyRequest error(%v)", err)
			w.WriteHeader(http.StatusInternalServerError)
			errRes := &NotifyRes{
				Code: "FAIL",
				Msg:  err.Error(),
			}
			errByte, err := json.Marshal(errRes)
			if err != nil {
				fmt.Printf("Marshal errRes(%v) error(%v)", errRes, err)
				return
			}
			w.Write(errByte)
			return
		}

		myNotifyReq, err = getNativeReq2Mine(notifyReq)
		if err != nil {
			fmt.Printf("get naticeReq error(%v)", err)
			w.WriteHeader(http.StatusInternalServerError)
			errRes := &NotifyRes{
				Code: "FAIL",
				Msg:  err.Error(),
			}
			errByte, err := json.Marshal(errRes)
			if err != nil {
				fmt.Printf("Marshal errRes(%v) error(%v)", errRes, err)
				return
			}
			w.Write(errByte)
			return
		}
		fmt.Printf("decode resource: %v\n", myNotifyReq.Resource.Plaintext)
		if options.OnCallBack != nil {
			if err = options.OnCallBack(ctx, myNotifyReq, nativeReq); err != nil {
				fmt.Printf("options-> OnCallBack error(%v)", err)
				w.WriteHeader(http.StatusInternalServerError)
				errRes := &NotifyRes{
					Code: "FAIL",
					Msg:  err.Error(),
				}
				errByte, err := json.Marshal(errRes)
				if err != nil {
					fmt.Printf("Marshal errRes(%v) error(%v)", errRes, err)
					return
				}
				w.Write(errByte)
				return
			}
		}

		//接收成功
		nativeReq.IsHandle = true
		w.WriteHeader(http.StatusOK)
	}
}

func getNativeReq2Mine(notifyReq *notify.Request) (myNotifyReq *NotifyReq, err error) {
	if notifyReq == nil {
		return nil, errors.New("notifyReq can not be nil")
	}
	myNotifyReq.ID = notifyReq.ID
	myNotifyReq.CreateTime = notifyReq.CreateTime.String()
	myNotifyReq.EventType = notifyReq.EventType
	myNotifyReq.ResourceType = notifyReq.ResourceType
	myNotifyReq.Summary = notifyReq.Summary
	myNotifyReq.Resource.Algorithm = notifyReq.Resource.Algorithm
	myNotifyReq.Resource.Ciphertext = notifyReq.Resource.Ciphertext
	myNotifyReq.Resource.OriginalType = notifyReq.Resource.OriginalType
	myNotifyReq.Resource.Nonce = notifyReq.Resource.Nonce
	myNotifyReq.Resource.AssociatedData = notifyReq.Resource.AssociatedData
	myNotifyReq.Resource.Plaintext = notifyReq.Resource.Plaintext
	return
}
