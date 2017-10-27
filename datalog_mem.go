package main

import (
	"bufio"
	"fmt"
	"os"
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
