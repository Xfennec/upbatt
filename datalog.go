package main

import (
	"bufio"
	"os"
	"time"
)

const dataLogPath = "/var/lib/upbatt/data.log"

// DataLog test struct
type DataLog struct {
	file     *os.File
	writer   *bufio.Writer
	messages chan string
	suspends []time.Time
}

// NewDataLog test
func NewDataLog() (*DataLog, error) {
	var dl DataLog
	f, err := os.OpenFile(dataLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	dl.writer = bufio.NewWriter(f)
	dl.file = f
	dl.messages = make(chan string) // unbuffered, to keep sync
	go dl.pump()

	return &dl, nil
}

// internal message channel pump
func (dl *DataLog) pump() {
	for str := range dl.messages {
		dl.writer.WriteString(str)
		dl.writer.Flush()
	}
}

// AppendRaw test
func (dl *DataLog) AppendRaw(str string) {
	dl.messages <- str
}

// Append test
func (dl *DataLog) Append(str string) {
	dl.AppendRaw(time.Now().Format(time.RFC3339) + ";" + str + "\n")
}

// AddSuspendEvent test
func (dl *DataLog) AddSuspendEvent() {
	dl.suspends = append(dl.suspends, time.Now())
}

// AnySuspendEventBefore test
func (dl *DataLog) AnySuspendEventBefore(date time.Time, window time.Duration) bool {
	for _, suspend := range dl.suspends {
		if date.Sub(suspend) <= window {
			return true
		}
	}
	return false
}
