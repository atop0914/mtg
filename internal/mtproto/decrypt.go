package mtproto

import (
	"encoding/binary"
	"errors"
)

// Payload represents decoded MTPROTO message
type Payload struct {
	AuthKeyID int64
	MsgID     int64
	Data      []byte
}

// DecodeMessage decodes an MTPROTO message from raw bytes
func DecodeMessage(data []byte) (*Payload, error) {
	if len(data) < 20 {
		return nil, ErrInvalidMessage
	}

	authKeyID := int64(binary.LittleEndian.Uint64(data[0:8]))
	msgID := int64(binary.LittleEndian.Uint64(data[8:16]))
	msgData := data[16:]

	return &Payload{
		AuthKeyID: authKeyID,
		MsgID:     msgID,
		Data:      msgData,
	}, nil
}

// EncodeMessage encodes an MTPROTO message to raw bytes
func EncodeMessage(authKeyID, msgID int64, data []byte) []byte {
	result := make([]byte, 16+len(data))
	binary.LittleEndian.PutUint64(result[0:8], uint64(authKeyID))
	binary.LittleEndian.PutUint64(result[8:16], uint64(msgID))
	copy(result[16:], data)
	return result
}

var ErrInvalidMessage = errors.New("invalid message")
