## AliPay
# 调用AliPayCommit生成支付宝支付url

1. 通过NewAliPayReq-> 生成*AliPayReq
2. 需要准备 支付宝应用ID:appID, 商户私钥:privateKey, 支付宝公钥:aliPublicKey
3. 调用AliPayCommit生成支付url

# 调用RefundByAliPay发起退款请求

1. NewAliPayRefundReq-> 生成*AliPayRefundReq
2. 调用RefundByAliPay发起退款
3. 返回*AliPayRefundRsp

# 支付回调通知

1. NotifyHandle-> 返回http.HandlerFunc,并且将支付状态更新在传入的*AliPayReq
2. 处理http.HandlerFunc