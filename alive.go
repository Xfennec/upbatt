package main

import (
	"bytes"
	"os"
	"time"
)

const aliveFilePath = "/var/lib/upbatt/alive.dat"

// AliveSchedule test
func AliveSchedule(delay time.Duration, dl *DataLog) error {

	fd, err := os.OpenFile(aliveFilePath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}

	go func() {
		var buffer []byte
		buffer = make([]byte, 128)
		for {
			now := time.Now()

			// read old value
			fd.Seek(0, 0)
			n, _ := fd.Read(buffer)
			if n == len(buffer) { // we have a full 128 bytes buffer
				pos := bytes.IndexByte(buffer, 0) // with some 0 somewhere
				if pos > 0 {
					org := string(buffer[:pos]) // extract string
					date, err := time.Parse(time.RFC3339, org)
					if err == nil {
						// if it was more than delay*2 time ago, add a stop / start
						// 	stopped := strings.TrimSpace(string(data))
						diff := now.Sub(date)
						if diff > delay*2 {
							dl.AppendRaw(date.Format(time.RFC3339) + ";stop\n")
							dl.Append("start")
						}
					}
				}
			} else {
				// probably a first start
				dl.Append("start")
			}

			// overwrite with current time
			fd.Seek(0, 0)
			copy(buffer, now.Format(time.RFC3339))
			fd.Write(buffer)
			time.Sleep(delay)
		}
	}()
	return nil
}
