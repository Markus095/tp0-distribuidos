package common

import (
	"net"
	"fmt"
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
	_, err := c.Conn.Read(response)
	if err != nil {
		log.Errorf("action: receive_ack | result: fail | error: %v", err)
		return nil, err
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

func (c *ClientNetwork)ReceiveWinners() (bool, []byte, error) {
    response := make([]byte, 1024) // Use a larger buffer to accommodate the payload
    n, err := c.Conn.Read(response)
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