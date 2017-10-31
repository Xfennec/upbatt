package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

const dataLogPath = "/var/lib/upbatt/data.log"
const aliveFilePath = "/var/lib/upbatt/alive.dat"
const aliveDelay = 5 * time.Second

func main() {

	server := flag.Bool("server", false, "start server daemon")
	force := flag.Bool("force", false, "force client event if daemon is not running")
	battery := flag.String("battery", "", "battery name")

	flag.Parse()

	if *server == true {
		if err := upbattServer(); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
			os.Exit(2)
		}
	} else {
		if err := upbattClient(*battery, *force); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
			os.Exit(3)
		}
	}
}
