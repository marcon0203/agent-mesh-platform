package infrastructure

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
)

// apiKeyCipher 用 AES-GCM 对称加密 ModelProvider.APIKey 后再落库，
// 避免 api_key 明文出现在数据库里。密钥由调用方传入（来自 Config.ModelProviderEncKey，
// 16/24/32 字节的 hex 编码字符串，分别对应 AES-128/192/256），不在这里直接读环境变量——
// 统一由 config.go 的 LoadConfig 决定"配置文件 vs 环境变量"的优先级。
type apiKeyCipher struct {
	gcm cipher.AEAD
}

func newAPIKeyCipher(keyHex string) (*apiKeyCipher, error) {
	if keyHex == "" {
		return nil, errors.New("model_provider_enc_key is not set")
	}
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return nil, errors.New("MODEL_PROVIDER_ENC_KEY must be hex-encoded: " + err.Error())
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &apiKeyCipher{gcm: gcm}, nil
}

// Encrypt 输出 nonce||ciphertext 拼接后的字节切片，直接写入 VARBINARY 列。
func (c *apiKeyCipher) Encrypt(plaintext string) ([]byte, error) {
	nonce := make([]byte, c.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return c.gcm.Seal(nonce, nonce, []byte(plaintext), nil), nil
}

func (c *apiKeyCipher) Decrypt(ciphertext []byte) (string, error) {
	nonceSize := c.gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	nonce, encrypted := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := c.gcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
