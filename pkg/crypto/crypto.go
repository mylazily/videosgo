// Package crypto 提供 bcrypt 密码哈希和 XOR 加解密功能
package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"

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

// XORDecrypt 使用 XOR 对 Base64 编码的数据进行解密
func XORDecrypt(encoded, key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("解密密钥不能为空")
	}
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("Base64 解码失败: %w", err)
	}

	keyBytes := []byte(key)
	keyLen := len(keyBytes)

	result := make([]byte, len(data))
	for i := range data {
		result[i] = data[i] ^ keyBytes[i%keyLen]
	}

	return string(result), nil
}

// XOREncryptHex 使用 XOR 加密并返回 Hex 编码结果
func XOREncryptHex(data, key string) (string, error) {
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

	return hex.EncodeToString(result), nil
}

// XORDecryptHex 使用 XOR 解密 Hex 编码的数据
func XORDecryptHex(hexStr, key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("解密密钥不能为空")
	}
	data, err := hex.DecodeString(hexStr)
	if err != nil {
		return "", fmt.Errorf("Hex 解码失败: %w", err)
	}

	keyBytes := []byte(key)
	keyLen := len(keyBytes)

	result := make([]byte, len(data))
	for i := range data {
		result[i] = data[i] ^ keyBytes[i%keyLen]
	}

	return string(result), nil
}

// GenerateRandomKey 生成指定长度的随机密钥
func GenerateRandomKey(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("生成随机密钥失败: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// IsEncrypted 检查字符串是否为 Base64 编码的加密数据
func IsEncrypted(s string) bool {
	// 简单检查是否为有效的 Base64 字符串
	s = strings.TrimSpace(s)
	if len(s) == 0 || len(s)%4 != 0 {
		return false
	}
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}
