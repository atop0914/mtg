package mtproto

import (
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	key := make([]byte, 32)
	plaintext := []byte("Hello, MTPROTO!")

	ciphertext, err := Encrypt(key, plaintext)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	decrypted, err := Decrypt(key, ciphertext)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("Decrypted mismatch: got %s, want %s", decrypted, plaintext)
	}
}

func TestGenerateMessageID(t *testing.T) {
	id1 := GenerateMessageID()
	id2 := GenerateMessageID()

	if id1 == id2 {
		t.Error("Message IDs should be unique")
	}
}

func TestEncodeDecodeMessage(t *testing.T) {
	authKeyID := int64(12345)
	msgID := int64(67890)
	data := []byte("test data")

	encoded := EncodeMessage(authKeyID, msgID, data)

	decoded, err := DecodeMessage(encoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if decoded.AuthKeyID != authKeyID {
		t.Errorf("AuthKeyID mismatch: got %d, want %d", decoded.AuthKeyID, authKeyID)
	}

	if decoded.MsgID != msgID {
		t.Errorf("MsgID mismatch: got %d, want %d", decoded.MsgID, msgID)
	}

	if string(decoded.Data) != string(data) {
		t.Errorf("Data mismatch: got %s, want %s", decoded.Data, data)
	}
}
