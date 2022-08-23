package util

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"math/rand"
	"regexp"
	"time"
)

func RandStr(strLen int, isNum bool) string {
	var str = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0987654321"
	if isNum {
		str = "0123456789"
	}
	bytes := []byte(str)
	result := make([]byte, 0, strLen)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < strLen; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

func MD5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func Sha256(str string) string {
	h := sha256.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func IsEmail(email string) bool {
	reg := regexp.MustCompile(`^[a-zA-Z0-9_-]+@[a-zA-Z0-9_-]+(\.[a-zA-Z0-9_-]+)+$`)
	return reg.MatchString(email)
}

func IsMobile(mobile string) bool {
	reg := regexp.MustCompile(`^1[0-9]{10}$`)
	return reg.MatchString(mobile)
}

func IsIpv4(ip string) bool {
	ok, _ := regexp.MatchString(`^((2[0-4]\d|25[0-5]|[01]?\d\d?)\.){3}(2[0-4]\d|25[0-5]|[01]?\d\d?)$`, ip)
	return ok
}

func IsNumber(str string) bool {
	ok, _ := regexp.MatchString(`^[0-9]+$`, str)
	return ok
}
