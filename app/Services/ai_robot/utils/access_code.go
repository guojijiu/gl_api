package utils

import (
	"crypto/rand"
	"math/big"
)

// GenerateAccessCode 生成指定长度的随机访问码。
func GenerateAccessCode(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	if length <= 0 {
		length = 6
	}
	out := make([]byte, length)
	max := big.NewInt(int64(len(chars)))
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			out[i] = chars[i%len(chars)]
			continue
		}
		out[i] = chars[n.Int64()]
	}
	return string(out)
}
