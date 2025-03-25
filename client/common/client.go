package common

import (
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
	"bufio"
	"strings"
	"fmt"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")


// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
	BetsPerBatch        uint16
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
	var datasetPath = fmt.Sprintf("dataset-%s.csv", c.config.ID)
    // Open the dataset file
	file, err := os.Open(datasetPath)
	if err != nil {
		log.Errorf("action: open_dataset | result: fail | error: %v", err)
		return
	}
	defer file.Close()
 
	// Convert agency ID to uint32
	agencyID, err := strconv.ParseUint(c.config.ID, 10, 32)
	if err != nil {
		log.Errorf("action: parse_agency_id | result: fail | error: %v", err)
		return
	}
 
	// Read the dataset line by line
	scanner := bufio.NewScanner(file)
	var bets []Bet
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, ",")
		if len(fields) < 5 {
			log.Errorf("action: parse_line | result: fail | error: invalid_line_format | line: %v", line)
			continue
		}
 
		// Create a bet from the line
		bet := Bet{
			FirstName: fields[0],
			LastName:  fields[1],
			Document:  fields[2],
			Birthdate: fields[3],
		}
		number, err := strconv.ParseUint(fields[4], 10, 16)
		if err != nil {
			log.Errorf("action: parse_number | result: fail | error: %v | line: %v", err, line)
			continue
		}
		bet.Number = uint16(number)
		bets = append(bets, bet)
	}
 
	if err := scanner.Err(); err != nil {
		log.Errorf("action: read_dataset | result: fail | error: %v", err)
		return
	}
 
	// Encode bets using protocol
	message := EncodeBets(uint32(agencyID), bets)
	log.Infof("action: encoded_bets | result: success | client_id: %v | bets_count: %d", c.config.ID, len(bets))
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
}
