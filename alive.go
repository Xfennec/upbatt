package main

import (
	"bytes"
	"os"
	"sync"
	"time"
)

func aliveCheckPauseLoop(delay time.Duration, dl *DataLogWriter, fd *os.File, wg *sync.WaitGroup) {
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

			if wg != nil {
				wg.Done()
				wg = nil
			}

			time.Sleep(delay)
		}
	}()
}

// AliveSchedule test
// (will wait the first loop to finish)
func AliveSchedule(delay time.Duration, dl *DataLogWriter) error {

	fd, err := os.OpenFile(aliveFilePath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go aliveCheckPauseLoop(delay, dl, fd, &wg)

	wg.Wait()
	return nil
}
