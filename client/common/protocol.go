package common

import (
	"encoding/binary"
	"time"
)

const (
	MaxFirstNameLength = 64
	MaxLastNameLength  = 64
	MaxDocumentLength  = 32
	MessageHeaderSize = 8   // 1 byte clientID + 7 bytes for number of bets
	BetSize          = 172  // 64 + 64 + 32 + 8 + 2 = 172 bytes per bet (removed agency)
)

// Bet represents a single betting record
type Bet struct {
	Agency    uint16
	FirstName string
	LastName  string
	Document  string
	Birthdate time.Time
	Number    uint16
}

// EncodeBets converts a slice of bets into a binary message following the protocol
func EncodeBets(clientID uint8, bets []Bet) []byte {
	messageSize := MessageHeaderSize + (BetSize * len(bets))
	message := make([]byte, messageSize)
	
	// Write client ID (1 byte)
	message[0] = clientID
	
	// Write number of bets (7 bytes)
	binary.BigEndian.PutUint64(message[1:8], uint64(len(bets)))
	
	offset := MessageHeaderSize
	for _, bet := range bets {
		// Write FirstName (64 bytes, padded with nulls)
		copy(message[offset:offset+MaxFirstNameLength], make([]byte, MaxFirstNameLength))
		copy(message[offset:], []byte(bet.FirstName))
		offset += MaxFirstNameLength
		
		// Write LastName (64 bytes, padded with nulls)
		copy(message[offset:offset+MaxLastNameLength], make([]byte, MaxLastNameLength))
		copy(message[offset:], []byte(bet.LastName))
		offset += MaxLastNameLength
		
		// Write Document (32 bytes, padded with nulls)
		copy(message[offset:offset+MaxDocumentLength], make([]byte, MaxDocumentLength))
		copy(message[offset:], []byte(bet.Document))
		offset += MaxDocumentLength
		
		// Write birthdate as YYYYMMDD (8 bytes)
		copy(message[offset:], []byte(bet.Birthdate.Format("20060102")))
		offset += 8
		
		// Write number (2 bytes)
		binary.BigEndian.PutUint16(message[offset:offset+2], bet.Number)
		offset += 2
	}
	
	return message
}