package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
)

// GenerateSecureJWTSecret 生成安全的JWT密钥
func GenerateSecureJWTSecret() (string, error) {
	// 生成64字节的随机密钥（更安全）
	bytes := make([]byte, 64)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("生成随机密钥失败: %v", err)
	}

	// 转换为十六进制字符串
	secret := hex.EncodeToString(bytes)

	// 添加一些特殊字符增加复杂度
	specialChars := "!@#$%^&*()_+-=[]{}|;:,.<>?"
	secret = secret + string(specialChars[bytes[0]%uint8(len(specialChars))])

	return secret, nil
}

func main() {
	fmt.Println("=== JWT密钥生成工具 ===")
	fmt.Println("正在生成安全的JWT密钥...")

	secret, err := GenerateSecureJWTSecret()
	if err != nil {
		fmt.Printf("生成密钥失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("生成的JWT密钥: %s\n", secret)
	fmt.Printf("密钥长度: %d 字符\n", len(secret))
	fmt.Println("\n请将以下内容添加到您的环境变量中:")
	fmt.Printf("export JWT_SECRET=\"%s\"\n", secret)
	fmt.Println("\n或者添加到 .env 文件中:")
	fmt.Printf("JWT_SECRET=%s\n", secret)
	fmt.Println("\n注意: 请妥善保管此密钥，不要泄露给他人！")
}
