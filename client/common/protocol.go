package common

import (
	"encoding/binary"
	"fmt"
)

const (
	MessageHeaderSize     = 8
	AgencyNumberSize      = 4
	MessageTypeSize       = 2
	AmountOfBetsSize      = 2 
	MaxFirstNameLength    = 64
	MaxLastNameLength     = 64
	MaxDocumentLength     = 8
	MaxDateLength         = 8
	MaxBetCodeLength      = 2
	BetSize               = MaxFirstNameLength + MaxLastNameLength + MaxDocumentLength + MaxDateLength + MaxBetCodeLength
	BetsMessage 	      = 1
	NotificationMessage   = 2
	WinnersRequestMessage = 3
)

func EncodeBets(agencyNumber uint32, bets []Bet) []byte {
    if len(bets) == 0 {
        log.Errorf("action: encode_bets | result: fail | error: no bets to encode")
        return nil
    }

    messageSize := MessageHeaderSize + len(bets)*BetSize
    message := make([]byte, messageSize)

    binary.BigEndian.PutUint32(message[0:4], agencyNumber)
    binary.BigEndian.PutUint16(message[4:6], BetsMessage)
    binary.BigEndian.PutUint16(message[6:8], uint16(len(bets)))

    offset := MessageHeaderSize
    for _, bet := range bets {
        copy(message[offset:offset+MaxFirstNameLength], []byte(fmt.Sprintf("%-64s", bet.FirstName)))
        offset += MaxFirstNameLength

        copy(message[offset:offset+MaxLastNameLength], []byte(fmt.Sprintf("%-64s", bet.LastName)))
        offset += MaxLastNameLength

        copy(message[offset:offset+MaxDocumentLength], []byte(fmt.Sprintf("%-8s", bet.Document)))
        offset += MaxDocumentLength

        birthdate := bet.Birthdate[:4] + bet.Birthdate[5:7] + bet.Birthdate[8:]
        copy(message[offset:offset+MaxDateLength], []byte(fmt.Sprintf("%-8s", birthdate)))
        offset += MaxDateLength

        binary.BigEndian.PutUint16(message[offset:offset+MaxBetCodeLength], bet.Number)
        offset += MaxBetCodeLength
    }

    return message
}

func EncodeNotification(agencyNumber uint32) []byte {
    message := make([]byte, MessageHeaderSize)
    binary.BigEndian.PutUint32(message[0:4], agencyNumber)
    binary.BigEndian.PutUint16(message[4:6], NotificationMessage)
    binary.BigEndian.PutUint16(message[6:8], 0)
    return message
}

func EncodeWinnersRequest(agencyNumber uint32) []byte {
    message := make([]byte, MessageHeaderSize)
    binary.BigEndian.PutUint32(message[0:4], agencyNumber)
    binary.BigEndian.PutUint16(message[4:6], WinnersRequestMessage)
    binary.BigEndian.PutUint16(message[6:8], 0)
    return message
}

