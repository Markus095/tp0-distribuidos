package common

import (
	"os"
	"bufio"
	"strings"
	"strconv"
	"fmt"
)
type Bet struct {
	FirstName string
	LastName  string
	Document  string
	Birthdate string
	Number    uint16
}

func ReadDataset(clientID string) ([]Bet, error) {
	var datasetPath = fmt.Sprintf("dataset-%s.csv", clientID)
	file, err := os.Open(datasetPath)
	if err != nil {
		log.Errorf("action: open_dataset | result: fail | error: %v", err)
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var bets []Bet

	for scanner.Scan() {
		line := scanner.Text()
		bet, err := processLine(line)
		if err == nil {
			bets = append(bets, bet)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Errorf("action: read_dataset | result: fail | error: %v", err)
		return nil, err
	}

	return bets, nil
}

func processLine(line string) (Bet, error) {
	fields := strings.Split(line, ",")
	if len(fields) < 5 {
		log.Errorf("action: parse_line | result: fail | error: invalid_line_format | line: %v", line)
		return Bet{}, fmt.Errorf("invalid line format")
	}

	number, err := strconv.ParseUint(fields[4], 10, 16)
	if err != nil {
		log.Errorf("action: parse_number | result: fail | error: %v | line: %v", err, line)
		return Bet{}, err
	}

	return Bet{
		FirstName: fields[0],
		LastName:  fields[1],
		Document:  fields[2],
		Birthdate: fields[3],
		Number:    uint16(number),
	}, nil
}
