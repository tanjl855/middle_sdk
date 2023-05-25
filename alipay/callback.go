package alipay

import (
	"context"
	"fmt"
	"net/http"
)

const (
	TradeStatusWaitBuyerPay = "WAIT_BUYER_PAY" //（交易创建，等待买家付款）
	TradeStatusClosed       = "TRADE_CLOSED"   //（未付款交易超时关闭，或支付完成后全额退款）
	TradeStatusSuccess      = "TRADE_SUCCESS"  //（交易支付成功）
	TradeStatusFinished     = "TRADE_FINISHED" //（交易结束，不可退款）
)

type Option struct {
	OnCallBack func(context.Context) error
}

/*
aliPublicKey: 支付宝公钥
a: 支付宝支付请求struct(支付状态更新在此)
*/
func NotifyHandle(aliPublicKey string, ctx context.Context, options *Option, a *AliPayReq) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		if a.PayStatus == TradeStatusSuccess || a.PayStatus == TradeStatusFinished || a.PayStatus == TradeStatusClosed {
			fmt.Printf("NotifyHandle-> AliPay notify is handled.")
			return
		}
		if err := r.ParseForm(); err != nil {
			fmt.Printf("NotifyHandle->parseForm error:%v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// 获取请求参数
		formData := r.PostForm
		sign := formData.Get("sign")
		signType := formData.Get("sign_type")
		formData.Del("sign")
		formData.Del("sign_type")

		// 验证签名
		if !verifySign(formData, sign, signType, aliPublicKey) {
			fmt.Printf("NotifyHandle->验证签名失败！")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		//logic
		if options != nil && options.OnCallBack != nil {
			err := options.OnCallBack(ctx)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		var out_trade_no, trade_no, trade_status string
		out_trade_no = formData.Get("out_trade_no") //商户订单号
		trade_no = formData.Get("trade_no")         //支付宝交易号
		trade_status = formData.Get("trade_status") //状态

		//更新状态
		a.OutTradeNo = out_trade_no
		a.PayStatus = trade_status
		a.TradeNo = trade_no

		//通知支付宝
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}
}
