## WechatPay
# 调用NativeCommit进行支付预请求返回二维码code_url

1. 通过NewNativeReq-> 生成 *NativeReq
2. 需要准备应用ID(appId), 商户号(mchId), path(商户私钥的本地位置)
3. 调用NativeCommit生成支付二维码url

# 调用RefundCommit发起退款请求

需要传入商户私钥的路径path
1. NewRefundReq-> 生成 *RefundReq
2. 调用RefundCommit发起退款
3. 返回*RefundResp

# 支付回调通知

1. NotifyHandle-> 返回http.HandlerFunc,并且将支付状态更新在传入的*NotifyReq
2. 处理http.HandlerFunc