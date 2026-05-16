// Package crypto 提供 bcrypt 密码哈希和 XOR 加解密功能
package crypto

import (
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	// DefaultCost bcrypt 默认成本因子
	DefaultCost = 10
)

// HashPassword 使用 bcrypt 对密码进行哈希
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	if err != nil {
		return "", fmt.Errorf("密码哈希失败: %w", err)
	}
	return string(bytes), nil
}

// CheckPassword 验证密码是否匹配哈希
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// XOREncrypt 使用 XOR 对数据进行加密，返回 Base64 编码结果
func XOREncrypt(data, key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("加密密钥不能为空")
	}
	dataBytes := []byte(data)
	keyBytes := []byte(key)
	keyLen := len(keyBytes)

	result := make([]byte, len(dataBytes))
	for i := range dataBytes {
		result[i] = dataBytes[i] ^ keyBytes[i%keyLen]
	}

	return base64.StdEncoding.EncodeToString(result), nil
}
