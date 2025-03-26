package common

import (
	"os"
	"os/signal"
	"syscall"
	"sync"
	"time"
	"strconv"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
	BetsPerBatch  uint16
}

type Client struct {
	config ClientConfig
	net    ClientNetwork
	done   chan bool
	wg     sync.WaitGroup
}

func NewClient(config ClientConfig) *Client {
	return &Client{
		config: config,
		net:    ClientNetwork{},
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
	log.Info("action: cleanup | result: in_progress")
	c.net.CloseConnection()
	log.Info("action: cleanup | result: success")
}

func (c *Client) StartClientLoop() {
    c.setupSignalHandler()

    if err := c.net.CreateClientSocket(c.config.ServerAddress); err != nil {
        log.Errorf("action: create_client_socket | result: fail | error: %v", err)
        return
    }
    defer c.net.CloseConnection() // Ensure the connection is closed after the process finishes

    // Send all bets and finish the process
    c.wg.Add(1)
    go func() {
        defer c.wg.Done()
        c.sendAndReceiveMessage(1) // Send all bets in one go
    }()

    c.wg.Wait()
    log.Infof("action: client_finished | result: success | client_id: %v", c.config.ID)
}

func (c *Client) sendAndReceiveMessage(msgID int) {
	bets, err := ReadDataset(c.config.ID)
	if err != nil {
		log.Errorf("action: read_dataset | result: fail | error: %v", err)
		return
	}
	agencyID, err := strconv.ParseUint(c.config.ID, 10, 32)
	if err != nil {
		log.Errorf("action: parse_agency_id | result: fail | error: %v", err)
		return
	}
	for i := 0; i < len(bets); i += int(c.config.BetsPerBatch) {
		end := i + int(c.config.BetsPerBatch)
		if end > len(bets) {
			end = len(bets)
		}
		batch := bets[i:end]
		message := EncodeBets(uint32(agencyID), batch)
		if err := c.net.SendMessage(message); err != nil {
			return
		}
		_, err = c.net.ReceiveAck()
		if err != nil {
			return
		}
		log.Infof("action: batch_sent | result: success | batch_size: %d", len(batch))
	}
	
}
