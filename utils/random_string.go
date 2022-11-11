package utils

import (
	"math/rand"
	"time"
)

var letterRunes = []rune("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")

// GetRandomString 获取任意长度字符串
func GetRandomString(n int) string {
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)

	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[r.Intn(len(letterRunes))]
	}
	return string(b)
}
