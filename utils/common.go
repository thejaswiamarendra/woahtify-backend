package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"

)

func GetEnv(key string) (string, error) {
	if val, exists := os.LookupEnv(key); exists {
		return val, nil
	}
	return "", fmt.Errorf("Env variable %s not found", key)
}

func GenerateSecureRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes)[:length], nil
}

func Encode(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func Decode(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

func Encrypt(plainText, key string) (string, error) {
	keyBytes := []byte(key)
	if len(keyBytes) != 16 && len(keyBytes) != 24 && len(keyBytes) != 32 {
		return "", fmt.Errorf("invalid AES key size: must be 16, 24, or 32 bytes")
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}

	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	cfb := cipher.NewCFBEncrypter(block, iv)
	cipherText := make([]byte, len(plainText))
	cfb.XORKeyStream(cipherText, []byte(plainText))

	final := append(iv, cipherText...)
	return Encode(final), nil
}

func Decrypt(encrypted, key string) (string, error) {
	keyBytes := []byte(key)
	if len(keyBytes) != 16 && len(keyBytes) != 24 && len(keyBytes) != 32 {
		return "", fmt.Errorf("invalid AES key size: must be 16, 24, or 32 bytes")
	}

	cipherTextWithIV, err := Decode(encrypted)
	if err != nil {
		return "", err
	}

	if len(cipherTextWithIV) < aes.BlockSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	iv := cipherTextWithIV[:aes.BlockSize]
	cipherText := cipherTextWithIV[aes.BlockSize:]

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}

	cfb := cipher.NewCFBDecrypter(block, iv)
	plainText := make([]byte, len(cipherText))
	cfb.XORKeyStream(plainText, cipherText)

	return string(plainText), nil
}
