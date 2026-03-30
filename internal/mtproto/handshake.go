package mtproto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
)

// GenerateKeyPair generates a new RSA key pair for handshake
func GenerateKeyPair() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 2048)
}

// HandshakeResult contains the result of a successful handshake
type HandshakeResult struct {
	AuthKey    []byte
	AuthKeyID  int64
	ServerSalt int64
}

// PerformHandshake performs the MTPROTO handshake process
func PerformHandshake(serverPublicKey *rsa.PublicKey) (*HandshakeResult, error) {
	// Generate nonce
	nonce := make([]byte, 32)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	// Encrypt nonce with server public key (simplified - just for demo)
	_, err := rsa.EncryptPKCS1v15(rand.Reader, serverPublicKey, nonce)
	if err != nil {
		return nil, err
	}

	// Derive auth key (simplified - real implementation uses DH)
	authKey := sha256.Sum256(nonce)
	authKeyID := int64(bigEndian(hash(authKey[:])))

	return &HandshakeResult{
		AuthKey:    authKey[:],
		AuthKeyID:  authKeyID,
		ServerSalt: 0, // Will be exchanged later
	}, nil
}

func bigEndian(b []byte) uint64 {
	var n uint64
	for i, v := range b {
		n = n<<8 | uint64(v)
		if i >= 7 {
			break
		}
	}
	return n
}

func hash(data []byte) []byte {
	h := sha256.Sum256(data)
	return h[:]
}
