package common

import (
	"os"
	"bufio"
	"strings"
	"strconv"
	"fmt"
	"encoding/binary"
)
type Bet struct {
	FirstName string
	LastName  string
	Document  string
	Birthdate string
	Number    uint16
}

func ReadDataset(clientID string) ([]Bet, error) {
    var datasetPath = fmt.Sprintf("data/agency-%s.csv", clientID)
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

    if len(bets) == 0 {
        log.Errorf("action: read_dataset | result: fail | error: no bets found")
        return nil, fmt.Errorf("no bets found in dataset")
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

func DecodeWinners(winnerBytes []byte) ([]uint16, error) {
    if len(winnerBytes)%2 != 0 {
        return nil, fmt.Errorf("invalid winners payload size")
    }

    var winners []uint16
    for i := 0; i < len(winnerBytes); i += 2 {
        winner := binary.BigEndian.Uint16(winnerBytes[i : i+2])
        winners = append(winners, winner)
    }

    return winners, nil
}