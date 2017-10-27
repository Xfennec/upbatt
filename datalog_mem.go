package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

type dataLogLine struct {
	time       time.Time
	eventName  string
	nativePath string
	data       map[string]string
}

// DataLogMem store a full DataLog in memory
type DataLogMem struct {
	filename string
	lines    []dataLogLine
}

// DataLogMemNew will parse filename to create a new DataLogMem
func DataLogMemNew(filename string) (*DataLogMem, error) {
	var dlm DataLogMem
	dlm.filename = filename

	fd, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(fd)

	for scanner.Scan() {
		var line dataLogLine
		fields := strings.Split(scanner.Text(), ";")
		time, err := time.Parse(time.RFC3339, fields[0])
		if err != nil {
			return nil, fmt.Errorf("can't parse date for line: %s", scanner.Text())
		}
		line.time = time
		line.eventName = fields[1]
		switch line.eventName {
		case "data":
			line.nativePath = fields[2]
			line.data = make(map[string]string)
			dataArray := strings.Split(fields[3], ",")
			for _, dataField := range dataArray {
				keyVal := strings.Split(dataField, "=")
				line.data[keyVal[0]] = keyVal[1]
			}
		case "start":
		case "stop":
		case "sleep":
		case "resume":
		case "online":
		case "offline":
		default:
			fmt.Fprintf(os.Stderr, "WARN: unknown event '%s' in %s\n", line.eventName, filename)
			continue
		}
		dlm.lines = append(dlm.lines, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return &dlm, nil
}
