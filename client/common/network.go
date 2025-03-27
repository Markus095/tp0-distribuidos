package common

import (
	"net"
	"fmt"
    "encoding/binary"
)

type ClientNetwork struct {
	Conn net.Conn
}

func (c *ClientNetwork) CreateClientSocket(serverAddr string) error {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		log.Criticalf("action: connect | result: fail | error: %v", err)
		return err
	}
	c.Conn = conn
	return nil
}

func (c *ClientNetwork) SendMessage(message []byte) error {
	_, err := c.Conn.Write(message)
	if err != nil {
		log.Errorf("action: send_message | result: fail | error: %v", err)
	}
	return err
}

func (c *ClientNetwork) CloseConnection() {
	if c.Conn != nil {
		c.Conn.Close()
	}
}


const (
	AnswerHeaderSize    = 4
	AnswerTypeSize      = 2
	AmountOfWinnersSize = 2
	ACKAnswer           = 1
	NoWinnersAnswer     = 2
	WinnersAnswer       = 3
)

func (c *ClientNetwork) ReceiveACK() (bool, error) {
	response := make([]byte, AnswerHeaderSize) 
	_, err := c.Conn.Read(response)
	if err != nil {
		log.Errorf("action: receive_ack | result: fail | error: %v", err)
		return false, err
	}

	answerType, _, err := DecodeAnswerType(response)
	if err != nil {
		log.Errorf("action: decode_ack | result: fail | error: %v", err)
		return false, err
	}

	if answerType != ACKAnswer {
		return false, fmt.Errorf("invalid answer type: expected %d, got %d", ACKAnswer, answerType)
	}
	return true, nil
}

func (c *ClientNetwork) ReceiveWinners() (bool, []byte, error) {
	response := make([]byte, AnswerHeaderSize) 
	_, err := c.Conn.Read(response)
	if err != nil {
		log.Errorf("action: receive_winners | result: fail | error: %v", err)
		return false, nil, err
	}

	answerType, payload, err := DecodeAnswerType(response)
	if err != nil {
		log.Errorf("action: decode_winners | result: fail | error: %v", err)
		return false, nil, err
	}
    log.Infof("action: receive_winners | result: success | answer_type: %d", answerType)
	if answerType == NoWinnersAnswer {
		log.Infof("action: receive_winners | result: success | info: no winners yet")
		return true, nil, nil
	} else if answerType != WinnersAnswer {
		return false, nil, fmt.Errorf("invalid answer type: expected %d, got %d", WinnersAnswer, answerType)
	}

	totalRead := 0
	for totalRead < len(payload) {
		n, err := c.Conn.Read(payload[totalRead:])
		if err != nil {
			log.Errorf("action: receive_winners | result: fail | error: %v", err)
			return false, nil, err
		}
		totalRead += n
	}

	return true, payload, nil
}

func DecodeAnswerType(answer []byte) (uint16, []byte, error) {
	if len(answer) < AnswerHeaderSize {
		log.Errorf("action: decode_answer_type | result: fail | error: invalid message size, message size: %d", len(answer))
		return 0, nil, fmt.Errorf("invalid message size: expected at least %d bytes, got %d", AnswerHeaderSize, len(answer))
	}

	// Extract the answer type (2 bytes)
	answerType := binary.BigEndian.Uint16(answer[0:2])

	// Extract the payload length (2 bytes)
	payloadLength := binary.BigEndian.Uint16(answer[2:4])

	// Read the remaining payload if present
	payload := make([]byte, payloadLength)
	return answerType, payload, nil
}