package main

import (
	"bufio"
	"os"
	"time"
)

// DataLog test struct
type DataLog struct {
	File   *os.File
	Writer *bufio.Writer
	Chan   chan string
}

// NewDataLog test
func NewDataLog() (*DataLog, error) {
	var dl DataLog
	f, err := os.OpenFile("data.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	dl.Writer = bufio.NewWriter(f)
	dl.File = f
	dl.Chan = make(chan string) // unbuffered, to keep sync
	go dl.pump()

	return &dl, nil
}

// internal channel pump
func (dl *DataLog) pump() {
	for str := range dl.Chan {
		dl.Writer.WriteString(time.Now().Format(time.RFC3339) + ";")
		dl.Writer.WriteString(str + "\n")
		dl.Writer.Flush()
	}
}

// Append test
func (dl *DataLog) Append(str string) {
	dl.Chan <- str
}
