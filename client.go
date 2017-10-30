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
	itPower := DataLogIteratorNew(dlm)
	for itPower.Prev() {
		eventName := itPower.Value().EventName
		if eventName == online || eventName == offline {
			powerEvent = itPower.Value()
			break
		}
	}

	if powerEvent == nil {
		fmt.Printf("power state unknown, no available data (yet?)\n")
		return nil
	}

	var percentageEvent *DataLogLine
	itPerc := DataLogIteratorNew(dlm)
	for itPerc.Prev() {
		if itPerc.Value().NativePath == battery && itPerc.Value().HasData(percentage) {
			percentageEvent = itPerc.Value()
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
			fmt.Printf("    charged since %s (%s)\n", SinceFmt(stateEvent.Time), stateEvent.Time.Format("2006-01-02 15:04"))
		case Charging:
			fmt.Printf("    charging since %s (%s)\n", SinceFmt(stateEvent.Time), stateEvent.Time.Format("2006-01-02 15:04"))
		}
	case offline:
		const up = 1
		const stopped = 2
		const suspended = 3

		var previousTime = itPower.Value().Time
		var durationUp time.Duration
		var durationStopped time.Duration
		var durationSuspended time.Duration
		var restarts = 0
		var pauses = 0

		// start from powerEvent
		for itPower.Next() {
			switch itPower.Value().EventName {
			case stop:
				diff := itPower.Value().Time.Sub(previousTime)
				durationUp += diff
				previousTime = itPower.Value().Time
			case start:
				diff := itPower.Value().Time.Sub(previousTime)
				durationStopped += diff
				restarts++
				previousTime = itPower.Value().Time
			case sleep:
				diff := itPower.Value().Time.Sub(previousTime)
				durationUp += diff
				previousTime = itPower.Value().Time
			case resume:
				diff := itPower.Value().Time.Sub(previousTime)
				durationSuspended += diff
				pauses++
				previousTime = itPower.Value().Time
			}
		}
		durationUp += time.Now().Sub(previousTime)

		fmt.Printf("On battery since %s\n", DurationFmt(durationUp))
		fmt.Printf("    + %s stopped (%d %s)\n", DurationFmt(durationStopped), restarts, Decline("restart", restarts))
		fmt.Printf("    + %s suspended (%d %s)\n", DurationFmt(durationSuspended), pauses, Decline("pause", pauses))

		var rateEvent *DataLogLine
		var tteEvent *DataLogLine
		itPerc.Prev() // it may be on the percentage line itself
		for itPerc.Next() {
			if itPerc.Value().NativePath == battery && itPerc.Value().HasData(rate) {
				rateEvent = itPerc.Value()
			}
			if itPerc.Value().NativePath == battery && itPerc.Value().HasData(timeToEmpty) {
				tteEvent = itPerc.Value()
			}
		}
		fmt.Printf("%s: %s%%", battery, strconv.FormatFloat(percentageEvent.GetDataPercentage(), 'f', -1, 64))
		if rateEvent != nil {
			fmt.Printf(", rate %.1f W", rateEvent.GetDataRate())
		}
		if tteEvent != nil {
			fmt.Printf(", time to empty %s", rateEvent.GetDataTte())
		}
		fmt.Printf("\n")

	}

	return nil
}
