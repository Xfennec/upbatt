package main

import (
	"flag"
	"fmt"
	"os"
	"time"
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

	if err := AliveSchedule(5*time.Second, datalog); err != nil {
		return err
	}

	fmt.Println("Server ready an running.")
	if err := SignalPump(ch, datalog); err != nil {
		return err
	}

	return nil
}

func main() {

	server := flag.Bool("server", false, "start server daemon")

	flag.Parse()

	if *server == true {
		if err := upbattServer(); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
			os.Exit(2)
		}
	} else {
		fmt.Println("We're the client.")
	}
}
