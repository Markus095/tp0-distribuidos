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

	for msgID := 1; msgID <= c.config.LoopAmount; msgID++ {
		select {
		case <-c.done:
			log.Info("action: client_loop | result: received_shutdown_signal")
			return
		default:
			if err := c.net.CreateClientSocket(c.config.ServerAddress); err != nil {
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
	message := EncodeBets(uint32(agencyID), bets)
	if err := c.net.SendMessage(message); err != nil {
		return
	}

	_, err = c.net.ReceiveAck()
	if err != nil {
		return
	}
}
