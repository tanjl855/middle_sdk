package kxsmsapi

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// 2.天翼云短信服务

/*
版本:3.0
编码:UTF-8
HTTP提交方式:POST
数据格式:JSON
参数名称和参数说明中规定的固定值必须与列表中完全一致（大小写敏感）
签名算法
签名方式：MD5
签名返回：32 位大写
按照以下示例参数顺序 加上密钥进行 MD5 加密
示例：
MD5(enterprise_no+account+timestamp+http_sign_Key)
*/

const TIANYIYUN = "TianYiyun"

type TianYiyunAdaptor struct {
	Sign         string `json:"sign"`          //签名
	EnterpriseNo string `json:"enterprise_no"` //企业编号
	Account      string `json:"account"`       //http接入账号
	Phones       string `json:"phones"`        //手机号码
	Content      string `json:"content"`       //短信内容
	TimeStamp    string `json:"timestamp"`     //时间戳,格式：yyyyMMddHHmmssSSS 用于生成Sign 20230417163200000
	Ip           string //http://ip/json/submit
	Debug        bool
}

type TianYiyunRes struct {
	Result    string `json:"result"`    //0成功 -1000失败
	Desc      string `json:"desc"`      //状态描述
	TimeStamp string `json:"timestamp"` //时间戳
	MsgId     string `json:"msgid"`     //消息标识，对应状态报告
	Sign      string `json:"sign"`      //签名
}

// Content示例："您的短信验证码:%v" 需要占位符,传入code作为验证码
// Sign需要大写32位
/*
ip:ip地址
enterpriseNo:企业编号
account:http接入账号
httpSignKey:密钥
content:短信内容
code:短信具体code
*/
func NewTianYiyunAdaptor(ip, enterpriseNo, account, httpSignKey, content, code string) *TianYiyunAdaptor {
	timestamp := GetTimeStamp(time.Now().Unix())
	signStr, err := CreateSign(enterpriseNo + account + timestamp + httpSignKey)
	if err != nil {
		fmt.Printf("NewTianYiyunAdaptor-> create sign error(%v) by md5", err)
		return nil
	}
	return &TianYiyunAdaptor{
		Ip:           ip,
		Sign:         signStr,
		EnterpriseNo: enterpriseNo,
		Account:      account,
		Content:      fmt.Sprintf(content, code),
		TimeStamp:    timestamp,
	}
}

func (adaptor *TianYiyunAdaptor) Send(phone string) (*SmsResponse, error) {
	Debug(adaptor.Debug, "Send message by TianYiyun(%v)", adaptor)
	//init
	adaptor.Phones = phone
	contentType := "application/json"
	url := fmt.Sprintf("http://%v/json/submit", adaptor.Ip)
	byteAdaptor, err := json.Marshal(adaptor)
	if err != nil {
		fmt.Printf("Send-> Marshal adaptor to byte slice error(%v)", err)
		return nil, err
	}
	buffer := bytes.NewBuffer(byteAdaptor)

	Debug(adaptor.Debug, "Init done")

	response, err := http.Post(url, contentType, buffer)
	if err != nil {
		fmt.Printf("Send-> post http://%v/json/submit error(%v)", adaptor.Ip, err)
		return nil, err
	}
	defer response.Body.Close()

	//读取resonse并且Unmarshal->TianYiyunRes
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("Send-> read response(%v) body error(%v)", response, err)
		return nil, err
	}
	tianRes := &TianYiyunRes{}
	if err = json.Unmarshal(body, tianRes); err != nil {
		fmt.Printf("Send-> unmarshal body(%v) to tianRes error(%v)", body, err)
		return nil, err
	}
	Debug(adaptor.Debug, "Send message success,tianYiyunRes: %v", string(body))

	//提取成SmsResponse
	smsRes := &SmsResponse{}
	smsRes.Code = GetCodeByDesc(tianRes.Desc)
	smsRes.Message = tianRes.Desc
	smsRes.TimeStamp = tianRes.TimeStamp
	smsRes.Sign = tianRes.Sign
	smsRes.RequestId = tianRes.MsgId
	smsRes.SmsType = TIANYIYUN
	Debug(adaptor.Debug, "Send message success,smsRes: %v", smsRes)
	return smsRes, err
}

// [md5]依据天翼云文档生成Sign 默认32位
func CreateSign(str string) (res string, err error) {
	hash := md5.New()
	_, err = hash.Write([]byte(str))
	if err != nil {
		fmt.Printf("NewTianYiyunAdaptor-> create sign from md5 error(%v)", err)
		return "", err
	}
	res = fmt.Sprintf("%x", hash.Sum(nil))
	res = strings.ToUpper(res) //32位大写
	return
}

func GetCodeByDesc(desc string) string {
	if desc == "" {
		return "desc为空"
	}
	//示例: "成功" -> "OK"
	if desc == "成功" {
		desc = "OK"
	}
	return desc
}

// 将int64的时间戳转换成string的yyyyMMddHHmmssSSS
func GetTimeStamp(timestamp int64) string {
	t := time.Unix(timestamp, 0)
	formatTime := t.Format("2006010215040512")
	return formatTime
}
