package common

import (
	"encoding/binary"
)

const (
	MessageHeaderSize  = 20  
	MaxFirstNameLength = 64
	MaxLastNameLength  = 64
	MaxDocumentLength  = 32
	MaxDateLength  = 8
	MaxBetCodeLength = 8
	BetSize = MaxFirstNameLength + MaxLastNameLength + MaxDocumentLength + MaxDateLength + MaxBetCodeLength
)

// Bet represents a single betting record
type Bet struct {
	FirstName string
	LastName  string
	Document  string
	Birthdate string
	Number    uint16
}

// EncodeBets converts a slice of bets into a binary message following the protocol
func EncodeBets(agencyNumber uint32, bets []Bet) []byte {
	// Calculate the total size of the message
	totalBets := uint32(len(bets))
	messageSize := MessageHeaderSize + int(totalBets)*BetSize

	// Create a byte slice to hold the message
	message := make([]byte, messageSize)

	// Write the agency number (4 bytes)
	binary.BigEndian.PutUint32(message[0:4], agencyNumber)

	// Write the number of bets (16 bytes, padded with zeros)
	binary.BigEndian.PutUint64(message[4:12], uint64(totalBets))

	// Encode each bet
	offset := MessageHeaderSize
	for _, bet := range bets {
		// Encode FirstName (MaxFirstNameLength bytes, padded with zeros)
		copy(message[offset:offset+MaxFirstNameLength], []byte(bet.FirstName))
		offset += MaxFirstNameLength

		// Encode LastName (MaxLastNameLength bytes, padded with zeros)
		copy(message[offset:offset+MaxLastNameLength], []byte(bet.LastName))
		offset += MaxLastNameLength

		// Encode Document (MaxDocumentLength bytes, padded with zeros)
		copy(message[offset:offset+MaxDocumentLength], []byte(bet.Document))
		offset += MaxDocumentLength

		// Encode Birthdate (MaxDateLength bytes, padded with zeros)
		copy(message[offset:offset+MaxDateLength], []byte(bet.Birthdate))
		offset += MaxDateLength

		// Encode Number (2 bytes)
		binary.BigEndian.PutUint16(message[offset:offset+MaxBetCodeLength], bet.Number)
		offset += MaxBetCodeLength
	}

	return message
}