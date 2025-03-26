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
	MaxDocumentLength     = 32
	MaxDateLength         = 8
	MaxBetCodeLength      = 2
	BetSize               = MaxFirstNameLength + MaxLastNameLength + MaxDocumentLength + MaxDateLength + MaxBetCodeLength
	BetsMessage 	      = 1
	NotificationMessage   = 2
	WinnersRequestMessage = 3
)

func EncodeBets(agencyNumber uint32, bets []Bet) []byte {
	messageSize := MessageHeaderSize + len(bets)*BetSize
	message := make([]byte, messageSize)

	binary.BigEndian.PutUint32(message[0:4], agencyNumber)
	binary.BigEndian.PutUint32(message[4:6], BetsMessage)
	binary.BigEndian.PutUint16(message[6:8], uint16(len(bets)))

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

func EncodeNotification(agencyNumber uint32) []byte {
    message := make([]byte, MessageHeaderSize)
    binary.BigEndian.PutUint32(message[0:4], agencyNumber)
    binary.BigEndian.PutUint16(message[4:6], NotificationMessage)
    binary.BigEndian.PutUint16(message[6:8], 0) // No batch size for notifications
    return message
}

func EncodeWinnersRequest(agencyNumber uint32) []byte {
    message := make([]byte, MessageHeaderSize)
    binary.BigEndian.PutUint32(message[0:4], agencyNumber)
    binary.BigEndian.PutUint16(message[4:6], WinnersRequestMessage)
    binary.BigEndian.PutUint16(message[6:8], 0) // No batch size for winners request
    return message
}

const (
	AnswerHeaderSize     = 4
	AnswerTypeSize       = 2
	AmountOfWinnersSize  = 2
	ACKAnswer            = 1
	WinnersAnswer        = 2
)

func DecodeAnswerType(answer []byte) (uint16, []byte, error) {
    if len(answer) < AnswerHeaderSize {
        return 0, nil, fmt.Errorf("invalid message size: expected at least %d bytes, got %d", AnswerHeaderSize, len(answer))
    }

    // Extract the answer type (2 bytes)
    answerType := binary.BigEndian.Uint16(answer[0:2])

    // Extract the payload length (2 bytes)
    payloadLength := binary.BigEndian.Uint16(answer[2:4])

    // Check if the payload length is valid
    if len(answer) < int(AnswerHeaderSize+payloadLength) {
        return 0, nil, fmt.Errorf("invalid payload size: expected %d bytes, got %d", payloadLength, len(answer)-AnswerHeaderSize)
    }

    // Extract the payload (if any)
    payload := answer[AnswerHeaderSize : AnswerHeaderSize+payloadLength]

    return answerType, payload, nil
}

func ReceiveACK(conn net.Conn) (bool, error) {
	response := make([]byte, 2)
	_, err := conn.Read(response)
	if err != nil {
		log.Errorf("action: receive_ack | result: fail | error: %v", err)
		return False, err
	}
	let type,_,_ = DecodeAnswerType(response)
	if type != ACKAnswer {
		return False, fmt.Errorf("invalid answer type: expected %d, got %d", ACKAnswer, type)
	}
	return True, nil
}

func ReceiveWinners(conn net.Conn) (bool, []byte, error) {
    response := make([]byte, 1024) // Use a larger buffer to accommodate the payload
    n, err := conn.Read(response)
    if (err != nil) {
        log.Errorf("action: receive_winners | result: fail | error: %v", err)
        return false, nil, err
    }

    answerType, payload, err := DecodeAnswerType(response[:n])
    if err != nil {
        return false, nil, err
    }

    if answerType != WinnersAnswer {
        return false, nil, fmt.Errorf("invalid answer type: expected %d, got %d", WinnersAnswer, answerType)
    }

    return true, payload, nil
}