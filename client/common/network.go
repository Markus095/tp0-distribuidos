package common

import (
	"net"
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
