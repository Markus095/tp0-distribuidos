package common

import (
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
	FirstName     string
	LastName      string
	Document      string
	Birthdate     string
	Number        uint16
}

// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	conn   net.Conn
	done   chan bool
	wg     sync.WaitGroup
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	return &Client{
		config: config,
		done:   make(chan bool),
	}
}

func (c *Client) setupSignalHandler() {
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGTERM)

    go func() {
        <-sigChan
        log.Info("action: handle_signal | result: success")
        close(c.done)
        c.cleanup()
        c.wg.Wait()
        log.Info("action: shutdown_client | result: success")
        os.Exit(0)
    }()
}

func (c *Client) cleanup() {
	if c.conn != nil {
		log.Info("action: close_connection | result: in_progress")
		err := c.conn.Close()
		if err != nil {
			log.Error("action: close_connection | result: fail | error: %v", err)
		} else {
			log.Info("action: close_connection | result: success")
		}
	}
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
	}
	c.conn = conn
	return nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	 // There is an autoincremental msgID to identify every message sent
    // Messages if the message amount threshold has not been surpassed
	c.setupSignalHandler()
	
	for msgID := 1; msgID <= c.config.LoopAmount; msgID++ {
		select {
		case <-c.done:
			log.Info("action: client_loop | result: received_shutdown_signal")
			return
		default:
			if err := c.createClientSocket(); err != nil {
				return
			}
			
			c.wg.Add(1)
			go func(id int) {
				defer c.wg.Done()
				c.sendAndReceiveMessage(id)
			}(msgID)

			time.Sleep(c.config.LoopPeriod)
		}
	}
	
	c.wg.Wait()
	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}

func (c *Client) sendAndReceiveMessage(msgID int) {
    // Create bet from environment variables
    bet := Bet{
        FirstName: c.config.FirstName,
        LastName:  c.config.LastName,
        Document:  c.config.Document,
        Birthdate: c.config.Birthdate,
        Number:    c.config.Number,
    }

    // Convert agency ID to uint32
    agencyID, err := strconv.ParseUint(c.config.ID, 10, 32)
    if err != nil {
        log.Errorf("action: parse_agency_id | result: fail | error: %v", err)
        return
    }
	
    // Encode bet using protocol
    message := EncodeBets(uint32(agencyID), []Bet{bet})
    log.Infof("action: encoded_bets | result: success | client_id: %v", c.config.ID)
    // Send full message
    _, err = c.conn.Write(message)
    if err != nil {
        log.Errorf("action: send_bet | result: fail | error: %v", err)
        return
    }

    // Read acknowledgment
    response := make([]byte, 2)
    _, err = c.conn.Read(response)
    if err != nil {
        log.Errorf("action: receive_ack | result: fail | error: %v", err)
        return
    }

    log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v",
        c.config.Document,
        c.config.Number)
}
