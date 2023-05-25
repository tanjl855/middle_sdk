package alipay

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"net/url"
	"sort"
	"strings"
)

const (
	kPublicKeyPrefix = "-----BEGIN PUBLIC KEY-----"
	kPublicKeySuffix = "-----END PUBLIC KEY-----"

	kPKCS1Prefix = "-----BEGIN RSA PRIVATE KEY-----"
	KPKCS1Suffix = "-----END RSA PRIVATE KEY-----"

	kPKCS8Prefix = "-----BEGIN PRIVATE KEY-----"
	KPKCS8Suffix = "-----END PRIVATE KEY-----"
)

func ParsePKCS1PrivateKey(data []byte) (key *rsa.PrivateKey, err error) {
	var block *pem.Block
	block, _ = pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to load private key")
	}

	key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return key, err
}

func FormatPKCS1PrivateKey(raw string) []byte {
	raw = strings.Replace(raw, kPKCS8Prefix, "", 1)
	raw = strings.Replace(raw, KPKCS8Suffix, "", 1)
	return formatKey(raw, kPKCS1Prefix, KPKCS1Suffix, 64)
}

func formatKey(raw, prefix, suffix string, lineCount int) []byte {
	if raw == "" {
		return nil
	}
	raw = strings.Replace(raw, prefix, "", 1)
	raw = strings.Replace(raw, suffix, "", 1)
	raw = strings.Replace(raw, " ", "", -1)
	raw = strings.Replace(raw, "\n", "", -1)
	raw = strings.Replace(raw, "\r", "", -1)
	raw = strings.Replace(raw, "\t", "", -1)

	var sl = len(raw)
	var c = sl / lineCount
	if sl%lineCount > 0 {
		c = c + 1
	}

	var buf bytes.Buffer
	buf.WriteString(prefix + "\n")
	for i := 0; i < c; i++ {
		var b = i * lineCount
		var e = b + lineCount
		if e > sl {
			buf.WriteString(raw[b:])
		} else {
			buf.WriteString(raw[b:e])
		}
		buf.WriteString("\n")
	}
	buf.WriteString(suffix)
	return buf.Bytes()
}

// LoadAliPayPublicKey 加载支付宝公钥
func (t *Client) LoadAliPayPublicKey(aliPublicKey string) error {
	var pub *rsa.PublicKey
	var err error
	if len(aliPublicKey) < 0 {
		return errors.New("alipay: alipay public key not found")
	}
	pub, err = ParsePublicKey(FormatPublicKey(aliPublicKey))
	if err != nil {
		return err
	}
	t.mux.Lock()
	t.aliPublicCertSN = "alipay-public-key"
	t.aliPublicKeyList[t.aliPublicCertSN] = pub
	t.mux.Unlock()
	return nil
}

func ParsePublicKey(data []byte) (key *rsa.PublicKey, err error) {
	var block *pem.Block
	block, _ = pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to load public key")
	}

	var pubInterface interface{}
	pubInterface, err = x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	key, ok := pubInterface.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("failed to load public key")
	}
	return key, err
}

func FormatPublicKey(raw string) []byte {
	prefix, suffix, lineCount := kPublicKeyPrefix, kPublicKeySuffix, 64
	if raw == "" {
		return nil
	}
	raw = strings.Replace(raw, prefix, "", 1)
	raw = strings.Replace(raw, suffix, "", 1)
	raw = strings.Replace(raw, " ", "", -1)
	raw = strings.Replace(raw, "\n", "", -1)
	raw = strings.Replace(raw, "\r", "", -1)
	raw = strings.Replace(raw, "\t", "", -1)

	var sl = len(raw)
	var c = sl / lineCount
	if sl%lineCount > 0 {
		c = c + 1
	}

	var buf bytes.Buffer
	buf.WriteString(prefix + "\n")
	for i := 0; i < c; i++ {
		var b = i * lineCount
		var e = b + lineCount
		if e > sl {
			buf.WriteString(raw[b:])
		} else {
			buf.WriteString(raw[b:e])
		}
		buf.WriteString("\n")
	}
	buf.WriteString(suffix)
	return buf.Bytes()
}

func ShaSign(data string, privateKey *rsa.PrivateKey) (string, error) {
	var hash = sha256.New()
	hash.Write([]byte(data))
	bys, err := rsa.SignPKCS1v15(nil, privateKey, crypto.SHA256, hash.Sum(nil))
	if err == nil {
		return base64.StdEncoding.EncodeToString(bys), nil
	} else {
		return "", err
	}
}

// 验证签名
func verifySign(formData url.Values, sign string, signType, alipayPublicKey string) bool {
	if formData == nil {
		return false
	}
	if sign == "" || signType == "" || alipayPublicKey == "" {
		return false
	}
	// 将参数按字典序排序
	keys := make([]string, 0, len(formData))
	for k := range formData {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 拼接待验签字符串
	var signData string
	for _, k := range keys {
		vs := formData[k]
		// 将参数值按升序排列
		sort.Strings(vs)
		for _, v := range vs {
			signData += k + "=" + v + "&"
		}
	}
	// 去掉最后一个&
	signData = signData[:len(signData)-1]

	// 读取支付宝公钥
	block, _ := pem.Decode([]byte(alipayPublicKey))
	if block == nil {
		return false
	}
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return false
	}

	// base64解码签名
	signBytes, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return false
	}

	// 根据签名类型选择哈希算法
	var hash crypto.Hash
	if signType == "RSA2" {
		hash = crypto.SHA256
	}

	// 计算待验签字符串的哈希值
	h := hash.New()
	h.Write([]byte(signData))
	hashed := h.Sum(nil)

	// 验证签名
	return rsa.VerifyPKCS1v15(pubKey.(*rsa.PublicKey), hash, hashed, signBytes) == nil
}
