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

func (c *ClientNetwork) ReceiveAck() ([]byte, error) {
    response := make([]byte, 2)
    totalRead := 0

    for totalRead < len(response) {
        n, err := c.Conn.Read(response[totalRead:])
        if err != nil {
            log.Errorf("action: receive_ack | result: fail | error: %v", err)
            return nil, err
        }
        totalRead += n
    }

    return response, nil
}

func (c *ClientNetwork) CloseConnection() {
	if c.Conn != nil {
		c.Conn.Close()
	}
}


func (c *ClientNetwork)ReceiveACK() (bool, error) {
    response := make([]byte, 4) 
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
    header := make([]byte, AnswerHeaderSize)
    totalRead := 0

    // Read the header first
    for totalRead < len(header) {
        n, err := c.Conn.Read(header[totalRead:])
        if err != nil {
            log.Errorf("action: receive_winners | result: fail | error: %v", err)
            return false, nil, err
        }
        totalRead += n
    }

    // Extract the payload length from the header (bytes 2-4)
    payloadLength := int(binary.BigEndian.Uint16(header[2:4]))

    // Handle empty payload case
    if payloadLength == 0 {
        log.Infof("action: receive_winners | result: success | info: no winners yet")
        return true, nil, nil
    }

    // Read the payload
    payload := make([]byte, payloadLength)
    totalRead = 0
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