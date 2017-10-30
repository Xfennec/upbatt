package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const online = "online"
const offline = "offline"
const data = "data"
const percentage = "percentage"
const state = "state"

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

func batteriesFromLog(dlm *DataLogMem) []string {
	battmap := make(map[string]bool)
	for it := DataLogIteratorNew(dlm); it.Prev(); {
		if it.Value().EventName == data {
			battmap[it.Value().NativePath] = true
		}
	}

	// map to array
	batteries := make([]string, 0)
	for key := range battmap {
		batteries = append(batteries, key)
	}
	return batteries
}

func upbattClient(battery string) error {
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

	batteries := batteriesFromLog(dlm)
	if len(batteries) == 0 {
		return errors.New("no battery information in log (yet?)")
	}
	if len(batteries) > 1 && battery == "" {
		return fmt.Errorf("multiple batteries detected in log, use -battery (%s)", strings.Join(batteries, ", "))
	}

	if battery == "" {
		battery = batteries[0]
	}

	var powerEvent *DataLogLine
	for it := DataLogIteratorNew(dlm); it.Prev(); {
		eventName := it.Value().EventName
		if eventName == online || eventName == offline {
			powerEvent = it.Value()
			break
		}
	}

	if powerEvent == nil {
		fmt.Printf("power state unknown, no available data (yet?)\n")
		return nil
	}

	var percentageEvent *DataLogLine
	for it := DataLogIteratorNew(dlm); it.Prev(); {
		if it.Value().NativePath == battery && it.Value().HasData(percentage) {
			percentageEvent = it.Value()
			break
		}
	}

	if percentageEvent == nil {
		fmt.Printf("battery percentage unknown, no available data (yet?)\n")
		return nil
	}

	// Percentage can't be more than 10 minutes BEFORE power event
	if powerEvent.Time.Sub(percentageEvent.Time) > 10*time.Minute {
		fmt.Printf("recent battery percentage unknown, no available data (yet?)\n")
		return nil
	}

	var stateEvent *DataLogLine
	for it := DataLogIteratorNew(dlm); it.Prev(); {
		if it.Value().NativePath == battery && it.Value().HasData(state) {
			stateEvent = it.Value()
			break
		}
	}

	if stateEvent == nil {
		fmt.Printf("battery state unknown, no available data (yet?)\n")
		return nil
	}

	switch powerEvent.EventName {
	case online:
		fmt.Printf("On line power since %s (%s)\n", SinceFmt(powerEvent.Time), powerEvent.Time.Format("2006-01-02 15:04"))
		fmt.Printf("%s: %s%%\n", battery, strconv.FormatFloat(percentageEvent.GetDataPercentage(), 'f', -1, 64))
		switch stateEvent.GetDataState() {
		case FullyCharged:
			fmt.Println("charged")
		case Charging:
			fmt.Println("charging")
		case Discharging:
			fmt.Println("discharging")
		}
	case offline:
		fmt.Printf("On battery since… who knows?\n")
	}

	return nil
}
