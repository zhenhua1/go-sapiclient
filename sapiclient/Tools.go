package sapiclient

import (
	"crypto/md5"
	"fmt"
	"strings"
)

// SEncryptSign
//
//	@Description: 生成加密串
//	@Author zzh 2023-10-30 15:25:26
//	@param appKey
//	@param appSecret
//	@param path
//	@param nonce
//	@param time
//	@return string
func SEncryptSign(appKey, appSecret, path, nonce, time string) string {
	sign := strings.ToUpper(Md5Encrypt(Md5Encrypt(appKey+strings.ToLower(path)+nonce+time) + appSecret))
	return sign
}

// Md5Encrypt
//
//	@Description: md5生成加密
//	@Author zzh 2023-11-01 11:24:38
//	@param data
//	@return string
func Md5Encrypt(data string) string {
	gmd5H := md5.New()
	gmd5H.Write([]byte(data))
	md5Val := fmt.Sprintf("%x", gmd5H.Sum(nil))
	return md5Val
}
