package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// DataLogLine holds information about a single log line
type DataLogLine struct {
	Time       time.Time
	EventName  string
	NativePath string
	Data       map[string]string
}

// DataLogMem store a full DataLog in memory
type DataLogMem struct {
	Filename string
	Lines    []DataLogLine
}

// HasData returns true if it's a data line with "name" field
func (line *DataLogLine) HasData(name string) bool {
	if line.EventName != data {
		return false
	}
	_, exists := line.Data[name]
	return exists
}

// GetDataState is an simple helper to get state
func (line *DataLogLine) GetDataState() int {
	val, _ := strconv.ParseInt(line.Data[state], 10, 32)
	return int(val)
}

// GetDataPercentage is an simple helper to get percentage
func (line *DataLogLine) GetDataPercentage() float64 {
	val, _ := strconv.ParseFloat(line.Data[percentage], 64)
	return val
}

// GetDataRate is an simple helper to get rate
func (line *DataLogLine) GetDataRate() float64 {
	val, _ := strconv.ParseFloat(line.Data[rate], 64)
	return val
}

// DataLogMemNew will parse filename to create a new DataLogMem
func DataLogMemNew(filename string) (*DataLogMem, error) {
	var dlm DataLogMem
	dlm.Filename = filename

	fd, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(fd)

	for scanner.Scan() {
		var line DataLogLine
		fields := strings.Split(scanner.Text(), ";")
		time, err := time.Parse(time.RFC3339, fields[0])
		if err != nil {
			return nil, fmt.Errorf("can't parse date for line: %s", scanner.Text())
		}
		line.Time = time
		line.EventName = fields[1]
		switch line.EventName {
		case "data":
			line.NativePath = fields[2]
			line.Data = make(map[string]string)
			dataArray := strings.Split(fields[3], ",")
			for _, dataField := range dataArray {
				keyVal := strings.Split(dataField, "=")
				line.Data[keyVal[0]] = keyVal[1]
			}
		case "start":
		case "stop":
		case "sleep":
		case "resume":
		case "online":
		case "offline":
		default:
			fmt.Fprintf(os.Stderr, "WARN: unknown event '%s' in %s\n", line.EventName, filename)
			continue
		}
		dlm.Lines = append(dlm.Lines, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return &dlm, nil
}
