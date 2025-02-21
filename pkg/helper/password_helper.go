package helper

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"errors"

	"github.com/spf13/viper"
)

// Using encryption instead of hashing because we want to reuse it for scraping
func EncryptPassword(viper *viper.Viper, plaintext string) (string, error) {
	salt := viper.GetString("APP_SALT_PASSWORD")
	key := viper.GetString("APP_KEY_PASSWORD")
	if salt == "" || key == "" {
		return "", errors.New("invalid .env file")
	}

	// Pad or truncate key to 32 bytes for AES-256
	paddedKey := make([]byte, 32)
	copy(paddedKey, []byte(key))

	block, err := aes.NewCipher(paddedKey)
	if err != nil {
		return "", err
	}

	// Generate IV from salt
	iv := make([]byte, aes.BlockSize)
	saltHasher := sha256.New()
	saltHasher.Write([]byte(salt))
	saltHash := saltHasher.Sum(nil)
	copy(iv, saltHash[:aes.BlockSize])

	// Pad plaintext to block size
	padding := aes.BlockSize - (len(plaintext) % aes.BlockSize)
	padtext := make([]byte, len(plaintext)+padding)
	copy(padtext, []byte(plaintext))
	for i := len(plaintext); i < len(padtext); i++ {
		padtext[i] = byte(padding)
	}

	// Encrypt
	ciphertext := make([]byte, len(padtext))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, padtext)

	// Encode as base64
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func DecryptPassword(viper *viper.Viper, encrypted string) (string, error) {
	salt := viper.GetString("APP_SALT_PASSWORD")
	key := viper.GetString("APP_KEY_PASSWORD")
	if salt == "" || key == "" {
		return "", errors.New("invalid .env file")
	}

	// Decode base64
	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	// Generate IV from salt
	iv := make([]byte, aes.BlockSize)
	saltHasher := sha256.New()
	saltHasher.Write([]byte(salt))
	saltHash := saltHasher.Sum(nil)
	copy(iv, saltHash[:aes.BlockSize])

	// Create cipher block
	paddedKey := make([]byte, 32)
	copy(paddedKey, []byte(key))

	block, err := aes.NewCipher(paddedKey)
	if err != nil {
		return "", err
	}

	// Decrypt
	plaintext := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plaintext, ciphertext)

	// Remove padding
	if len(plaintext) == 0 {
		return "", errors.New("plaintext is empty")
	}
	padding := int(plaintext[len(plaintext)-1])
	if padding > len(plaintext) {
		return "", errors.New("invalid padding")
	}
	unpadded := plaintext[:len(plaintext)-padding]

	return string(unpadded), nil
}

func CompareEncryptedPassword(viper *viper.Viper, encrypted, plaintext string) (bool, error) {
	encryptedPlaintext, err := EncryptPassword(viper, plaintext)
	if err != nil {
		return false, err
	}

	return encrypted == encryptedPlaintext, nil
}
