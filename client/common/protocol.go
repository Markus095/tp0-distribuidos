package common

import (
	"encoding/binary"
	"fmt"
)

const (
	MessageHeaderSize  = 6 
	MaxFirstNameLength = 64
	MaxLastNameLength  = 64
	MaxDocumentLength  = 32
	MaxDateLength      = 8
	MaxBetCodeLength   = 2
	BetSize            = MaxFirstNameLength + MaxLastNameLength + MaxDocumentLength + MaxDateLength + MaxBetCodeLength
)

// Bet represents a single betting record
type Bet struct {
	FirstName string
	LastName  string
	Document  string
	Birthdate string
	Number    uint16
}

func EncodeBets(agencyNumber uint32, bets []Bet) []byte {
	messageSize := MessageHeaderSize + len(bets)*BetSize
	message := make([]byte, messageSize)

	binary.BigEndian.PutUint32(message[0:4], agencyNumber)
	binary.BigEndian.PutUint16(message[4:6], uint16(len(bets)))

	offset := MessageHeaderSize
	for _, bet := range bets {
		copy(message[offset:offset+MaxFirstNameLength], []byte(fmt.Sprintf("%-64s", bet.FirstName)))
		offset += MaxFirstNameLength
		
		copy(message[offset:offset+MaxLastNameLength], []byte(fmt.Sprintf("%-64s", bet.LastName)))
		offset += MaxLastNameLength
		
		copy(message[offset:offset+MaxDocumentLength], []byte(fmt.Sprintf("%-32s", bet.Document)))
		offset += MaxDocumentLength
		
		birthdate := bet.Birthdate[:4] + bet.Birthdate[5:7] + bet.Birthdate[8:]
		copy(message[offset:offset+MaxDateLength], []byte(fmt.Sprintf("%-8s", birthdate)))
		offset += MaxDateLength

		binary.BigEndian.PutUint16(message[offset:offset+MaxBetCodeLength], bet.Number)
		offset += MaxBetCodeLength
	}

	return message
}
