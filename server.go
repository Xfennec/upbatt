package main

import (
	"fmt"
	"strconv"
)

// I'm not very proud of this one, need a rewrite
// get from log: online/offline state + battery state + battery percentage
// no error reporting if DataLogMemNew fails, it's not our job as a server
func getLogPowerAndBatteriesStats(statBatteries []BatteryStats) string {
	var powerLog = ""
	if dlm, err := DataLogMemNew(dataLogPath); err == nil {
		for it := DataLogIteratorNew(dlm); it.Prev(); {
			if it.Value().EventName == offline || it.Value().EventName == online {
				powerLog = it.Value().EventName
				break
			}
		}
		for key, statBattery := range statBatteries {
			for it := DataLogIteratorNew(dlm); it.Prev(); {
				if it.Value().NativePath == statBattery.NativePath && it.Value().HasData(percentage) {
					statBatteries[key].Log.Percentage = it.Value().GetDataPercentage()
					break
				}
			}
			for it := DataLogIteratorNew(dlm); it.Prev(); {
				if it.Value().NativePath == statBattery.NativePath && it.Value().HasData(state) {
					statBatteries[key].Log.State = it.Value().GetDataState()
					break
				}
			}
		}
	}
	return powerLog
}

func upbattServer() error {
	devices, err4 := EnumerateDevices()
	if err4 != nil {
		return err4
	}

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

	powerReal, err5 := GetRealPowerState(devices)
	if err5 != nil {
		return err5
	}

	statBatteries := GetRealBatteriesStats(devices)
	powerLog := getLogPowerAndBatteriesStats(statBatteries)

	datalog, err3 := NewDataLog()
	if err3 != nil {
		return err3
	}

	if err := AliveSchedule(aliveDelay, datalog); err != nil {
		return err
	}

	if powerReal != powerLog {
		fmt.Printf("power status mismatch, fixing (real=%s, log=%s)\n", powerReal, powerLog)
		datalog.Append(powerReal)
	}

	// reconciliation
	for _, statBattery := range statBatteries {
		if statBattery.Real.Percentage >= 0 {
			if statBattery.Log.Percentage != statBattery.Real.Percentage {
				fmt.Printf("%s: '%s' mismatch, fixing\n", statBattery.NativePath, percentage)
				datalog.Append(data + ";" + statBattery.NativePath + ";" + percentage + "=" + FloatFmt(statBattery.Real.Percentage))
			}
		}
		if statBattery.Real.State != -1 {
			if statBattery.Log.State != statBattery.Real.State {
				fmt.Printf("%s: '%s' mismatch, fixing\n", statBattery.NativePath, state)
				datalog.Append(data + ";" + statBattery.NativePath + ";" + state + "=" + strconv.Itoa(statBattery.Real.State))
			}
		}
	}

	fmt.Println("Server ready an running.")
	if err := SignalPump(ch, datalog); err != nil {
		return err
	}

	return nil
}
