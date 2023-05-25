## Aliyun & TianYiyun

需要在aliyun.go中配置aliyunKey进行使用 -> 后续需要进行优化

1. 通过NewAliyunAdaptor初始化阿里云请求结构、通过NewTianYiyunAdaptor初始化天翼云请求结构

2. 调用SendSms,通过参数...Adaptor控制发送短信平台的顺序

3. 如果第一个发送失败，就会调用第二个平台发送短信