package main

import (
	"bytes"
	"os"
	"time"
)

const aliveFilePath = "/var/lib/upbatt/alive.dat"

func aliveCheckPauseLoop(delay time.Duration, dl *DataLog, fd *os.File) {
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
						// if it was more than delay*2 time ago (and not
						// a proper suspend/resume) add a stop / start
						diff := now.Sub(date)
						if diff > delay*2 && dl.AnySuspendEventBefore(date, delay*2) == false {
							dl.AppendRaw(date.Format(time.RFC3339) + ";stop\n")
							dl.Append("start")
						}
					}
				}
			} else {
				// probably a first start (no existing alive.dat)
				dl.Append("start")
			}

			// overwrite with current time
			fd.Seek(0, 0)
			copy(buffer, now.Format(time.RFC3339))
			fd.Write(buffer)
			time.Sleep(delay)
		}
	}()
}

// AliveSchedule test
func AliveSchedule(delay time.Duration, dl *DataLog) error {

	fd, err := os.OpenFile(aliveFilePath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}

	go aliveCheckPauseLoop(delay, dl, fd)

	return nil
}
