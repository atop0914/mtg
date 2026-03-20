package mtproto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"
)

// Encrypt encrypts data using AES-256-CBC with random IV
func Encrypt(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)

	return ciphertext, nil
}

// Decrypt decrypts data using AES-256-CBC
func Decrypt(key, ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)

	return ciphertext, nil
}

// GenerateMessageID generates a unique message ID
func GenerateMessageID() int64 {
	// Current time in milliseconds with random lower bits
	id := (nowUnixMillis() << 32) | (randInt32() & 0xFFFFFFFF)
	return int64(id)
}

func nowUnixMillis() int64 {
	return 0 // Placeholder - use time.Now().UnixMilli()
}

func randInt32() int32 {
	return 0 // Placeholder - use rand.Int31()
}
