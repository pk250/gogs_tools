package ai

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// AIReviewer AI 代码审查适配器接口
type AIReviewer interface {
	Review(prompt, diff string) (string, error)
}

// New 根据 provider 名称返回对应适配器
func New(provider, apiKey, model string) (AIReviewer, error) {
	switch provider {
	case "claude":
		return &ClaudeReviewer{APIKey: apiKey, Model: model}, nil
	case "openai":
		return &OpenAIReviewer{APIKey: apiKey, Model: model}, nil
	default:
		return nil, errors.New("未知 AI 服务商: " + provider)
	}
}

// appKey 用于 AES-256-GCM 加密，32 字节
// 生产环境建议从环境变量或配置文件读取
var appKey = []byte("gogs_tools_aes_key_32bytes_pad!!")

// Encrypt 使用 AES-256-GCM 加密明文，返回 base64 编码密文
func Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(appKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密 Encrypt 产生的 base64 密文
func Decrypt(cipherB64 string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(cipherB64)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(appKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("密文过短")
	}
	plaintext, err := gcm.Open(nil, data[:nonceSize], data[nonceSize:], nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
