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

func DecodeWinners(winnerBytes []byte) ([]uint32, error) {
    if len(winnerBytes)%4 != 0 {
        log.Infof("action: decode_winners | result: success | info: no winners in payload")
        return []uint32{}, nil
    }

    var winners []uint32
    for i := 0; i < len(winnerBytes); i += 4 {
        winner := binary.BigEndian.Uint32(winnerBytes[i : i+4])
        winners = append(winners, winner)
    }

    return winners, nil
}