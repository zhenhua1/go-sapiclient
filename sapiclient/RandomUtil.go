package sapiclient

import (
	"crypto/sha1"
	"fmt"
	"github.com/syyongx/php2go"
	"math"
	"math/rand"
	"strconv"
	"strings"
)

// Alnum
//
//	@Description: 生成数字和字母
//	@Author zzh 2023-10-30 15:13:26
//	@param length
func Alnum(length ...int) (random string) {
	alnumLen := 6
	if len(length) > 0 {
		alnumLen = length[0]
	}
	return build("alnum", alnumLen)
}

// Alpha
//
//	@Description:仅生成字符
//	@Author zzh 2023-10-30 15:14:12
//	@param length
func Alpha(length ...int) (random string) {
	alnumLen := 6
	if len(length) > 0 {
		alnumLen = length[0]
	}
	return build("alpha", alnumLen)
}

// Numeric
//
//	@Description:生成指定长度的随机数字
//	@Author zzh 2023-10-30 15:14:54
//	@param length
func Numeric(length ...int) (random string) {
	alnumLen := 4
	if len(length) > 0 {
		alnumLen = length[0]
	}
	return build("numeric", alnumLen)
}

// Nozero
//
//	@Description:生成指定长度的无0随机数字
//	@Author zzh 2023-10-30 15:16:12
//	@param length
func Nozero(length ...int) (random string) {
	alnumLen := 4
	if len(length) > 0 {
		alnumLen = length[0]
	}
	return build("nozero", alnumLen)
}

// build
//
//	@Description: 构建生成随机数
//	@Author zzh 2023-10-31 18:07:23
//	@param randomType
//	@param length
//	@return random
func build(randomType string, length int) (random string) {
	poolStr := ""
	switch randomType {
	case "alpha", "alnum", "numeric", "nozero":
		switch randomType {
		case "alpha":
			poolStr = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
		case "alnum":
			poolStr = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
		case "numeric":
			poolStr = "0123456789"
		case "nozero":
			poolStr = "123456789"
		}
		random = php2go.Substr(shuffle(strings.Repeat(poolStr, int(math.Ceil(float64(length)/float64(len(poolStr)))))), 0, length)
	case "unique", "md5":
		random = Md5Encrypt(php2go.Uniqid(strconv.Itoa(randN(1000000000, 9999999999))))
	case "encrypt":
	case "sha1":
		sha1New := sha1.New()
		sha1New.Write([]byte(php2go.Uniqid(strconv.Itoa(randN(1000000000, 9999999999)))))
		random = fmt.Sprintf("%x", sha1New.Sum(nil))
	}
	return
}

// randN
//
//	@Description: 生成随机数据
//	@Author zzh 2023-11-01 10:24:21
//	@param min
//	@param max
//	@return int
func randN(min, max int) int {
	if min >= max {
		return min
	}
	if min >= 0 {
		return rand.Intn(max-min+1) + min
	}
	return rand.Intn(max+(0-min)+1) - (0 - min)
}

// shuffle
//
//	@Description: 对字符串进行随机排序
//	@Author zzh 2023-11-01 10:30:18
//	@param inputStr
//	@return string
func shuffle(inputStr string) string {
	runeStr := []rune(inputStr)
	for i := len(runeStr) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		runeStr[i], runeStr[j] = runeStr[j], runeStr[i]
	}
	shuffledStr := string(runeStr)
	return shuffledStr
}
