package mtproto

import (
	"testing"
)

func BenchmarkEncrypt(b *testing.B) {
	key := make([]byte, 32)
	plaintext := make([]byte, 1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Encrypt(key, plaintext)
	}
}

func BenchmarkDecrypt(b *testing.B) {
	key := make([]byte, 32)
	plaintext := make([]byte, 1024)
	ciphertext, _ := Encrypt(key, plaintext)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Decrypt(key, ciphertext)
	}
}
