package main

import (
	"io/ioutil"
	"os"
	"strings"
	"time"
)

const aliveFilePath = "/var/lib/upbatt/alive.log"

// AliveSchedule test
func AliveSchedule(delay time.Duration) error {

	fd, err := os.Create(aliveFilePath)
	if err != nil {
		return err
	}

	go func() {
		for {
			fd.Seek(0, 0)
			fd.Write([]byte(time.Now().Format(time.RFC3339) + "\n"))
			time.Sleep(delay)
		}
	}()
	return nil
}

// AliveCheck test
func AliveCheck(dl *DataLog) error {

	if _, err := os.Stat(aliveFilePath); os.IsNotExist(err) {
		return nil
	}

	data, err := ioutil.ReadFile(aliveFilePath)
	if err != nil {
		return err
	}
	stopped := strings.TrimSpace(string(data))
	dl.AppendRaw(stopped + ";stop\n")
	return nil
}
