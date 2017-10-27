package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {

	server := flag.Bool("server", false, "start server daemon")

	flag.Parse()

	if *server == true {
		if err := upbattServer(); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
			os.Exit(2)
		}
	} else {
		if err := upbattClient(); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
			os.Exit(3)
		}
	}
}
