package mtproto

import (
	"crypto/aes"
	"crypto/cipher"
	cryptorand "crypto/rand"
	"errors"
	"io"
	"math/rand"
	"time"
)

// Encrypt encrypts data using AES-256-CBC with random IV and PKCS7 padding
func Encrypt(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Apply PKCS7 padding
	padLen := aes.BlockSize - (len(plaintext) % aes.BlockSize)
	padded := make([]byte, len(plaintext)+padLen)
	copy(padded, plaintext)
	for i := len(plaintext); i < len(padded); i++ {
		padded[i] = byte(padLen)
	}

	ciphertext := make([]byte, aes.BlockSize+len(padded))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(cryptorand.Reader, iv); err != nil {
		return nil, err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], padded)

	return ciphertext, nil
}

// Decrypt decrypts data using AES-256-CBC and removes PKCS7 padding
func Decrypt(key, ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < aes.BlockSize {
		return nil, ErrCiphertextTooShort
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	iv := ciphertext[:aes.BlockSize]
	encrypted := ciphertext[aes.BlockSize:]

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(encrypted, encrypted)

	// Remove PKCS7 padding
	padLen := int(encrypted[len(encrypted)-1])
	if padLen > aes.BlockSize || padLen == 0 {
		return nil, errors.New("invalid padding")
	}
	decrypted := encrypted[:len(encrypted)-padLen]

	return decrypted, nil
}

// ErrCiphertextTooShort is returned when ciphertext is too short to decrypt
var ErrCiphertextTooShort = errors.New("ciphertext too short")

// GenerateMessageID generates a unique message ID
func GenerateMessageID() int64 {
	// Initialize seed once
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	// Current time in milliseconds with random lower bits
	id := (time.Now().UnixMilli() << 32) | (int64(rand.Int31()) & 0xFFFFFFFF)
	return id
}
