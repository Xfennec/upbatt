package main

import (
	"fmt"
)

func upbattServer() error {
	if err := Signal(); err != nil {
		return err
	}

	if err := SignalSystemd(); err != nil {
		return err
	}

	ch, err2 := Signals()
	if err2 != nil {
		return err2
	}

	datalog, err3 := NewDataLog()
	if err3 != nil {
		return err3
	}

	if err := AliveSchedule(aliveDelay, datalog); err != nil {
		return err
	}

	fmt.Println("Server ready an running.")
	if err := SignalPump(ch, datalog); err != nil {
		return err
	}

	return nil
}
