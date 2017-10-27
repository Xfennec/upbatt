package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"time"
)

// based on alive.go:aliveCheckPauseLoop() code, needs dedup
func readAliveDate(filename string) (*time.Time, error) {
	var buffer []byte
	buffer = make([]byte, 128)
	fd, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	n, _ := fd.Read(buffer)
	if n != len(buffer) {
		return nil, errors.New("can't read full alive block")
	}
	pos := bytes.IndexByte(buffer, 0)
	if pos < 0 {
		return nil, errors.New("alive block seems full of garbage")
	}
	org := string(buffer[:pos])
	date, err := time.Parse(time.RFC3339, org)
	if err != nil {
		return nil, errors.New("can't parse alive block date")
	}
	return &date, nil
}

func checkAliveDate(filename string, delay time.Duration) error {
	alive, err := readAliveDate(filename)
	if err != nil {
		return fmt.Errorf("%s (%s)", err, filename)
	}
	diff := time.Now().Sub(*alive)
	if diff > delay*2 {
		return fmt.Errorf("outdated alive block (%s)", filename)
	}
	return nil
}

func upbattClient() error {
	err := checkAliveDate(aliveFilePath, aliveDelay)
	if err != nil {
		return fmt.Errorf("it seems that upbatt server is not currently running\nreason: %s", err)
	}

	dlm, err := DataLogMemNew(dataLogPath)
	if err != nil {
		return fmt.Errorf("can't parse log %s: %s", dataLogPath, err)
	}

	if len(dlm.Lines) == 0 {
		return errors.New("empty log")
	}

	state := "unknown"
	for it := DataLogIteratorNew(dlm); it.Prev(); {
		event := it.Value().EventName
		if event == "online" || event == "offline" {
			state = event
			break
		}
	}

	fmt.Printf("state: %s\n", state)

	return nil
}
