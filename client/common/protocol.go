package common

import (
	"encoding/binary"
)

const (
	MaxAgencyIDLength  = 32  // New constant for agency ID length
	MaxFirstNameLength = 64
	MaxLastNameLength  = 64
	MaxDocumentLength  = 32
	MessageHeaderSize  = 8    // 1 byte for number of bets + 7 bytes reserved
	BetSize           = 204   // 32 + 64 + 64 + 32 + 8 + 4 = 204 bytes per bet
)

// Bet represents a single betting record
type Bet struct {
	Agency    string
	FirstName string
	LastName  string
	Document  string
	Birthdate string    // Changed from time.Time to string
	Number    uint16
}

// EncodeBets converts a slice of bets into a binary message following the protocol
func EncodeBets(bets []Bet) []byte {
	messageSize := MessageHeaderSize + (BetSize * len(bets))
	message := make([]byte, messageSize)
	
	// Write number of bets (8 bytes)
	binary.BigEndian.PutUint64(message[0:8], uint64(len(bets)))
	
	offset := MessageHeaderSize
	for _, bet := range bets {
		// Write Agency ID (32 bytes, padded with nulls)
		copy(message[offset:offset+MaxAgencyIDLength], make([]byte, MaxAgencyIDLength))
		copy(message[offset:], []byte(bet.Agency))
		offset += MaxAgencyIDLength
		
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
		
		// Write birthdate as string (8 bytes)
		copy(message[offset:offset+8], []byte(bet.Birthdate[:8])) // Assumes YYYYMMDD format
		offset += 8
		
		// Write number (2 bytes)
		binary.BigEndian.PutUint16(message[offset:offset+2], bet.Number)
		offset += 2
	}
	
	return message
}