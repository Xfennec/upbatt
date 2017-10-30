package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/godbus/dbus"
)

// Signal catches all signals.
func Signal() (err error) {

	conn, err := dbus.SystemBus()
	if err != nil {

		return
	}

	call := conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, "sender=org.freedesktop.UPower,type=signal")
	if call.Err != nil {

		return call.Err
	}

	return
}

// SignalSystemd catches all signals.
func SignalSystemd() (err error) {

	conn, err := dbus.SystemBus()
	if err != nil {

		return
	}

	call := conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, "sender=org.freedesktop.login1,type=signal")
	if call.Err != nil {

		return call.Err
	}

	return
}

// Signals returns a channel with all signals
func Signals() (ch chan *dbus.Signal, err error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, err
	}
	ch = make(chan *dbus.Signal, 10)
	conn.Signal(ch)
	return
}

// SignalPump test
func SignalPump(ch chan *dbus.Signal, datalog *DataLogWriter) error {

	for sig := range ch {

		// Discards UPower "composite" device because we're working on "real" devices
		if sig.Path == "/org/freedesktop/UPower/devices/DisplayDevice" {
			continue
		}

		switch sig.Name {
		case "org.freedesktop.DBus.Properties.PropertiesChanged":
			// PropertiesChanged's Body[]:
			// <arg type="s" name="interface_name"/>
			// <arg type="a{sv}" name="changed_properties"/>
			// <arg type="as" name="invalidated_properties"/>
			intf, castOk := sig.Body[0].(string)
			if castOk == false {
				fmt.Println("failed cast to string: ", sig)
				continue
			}

			changedProperties, castOk2 := sig.Body[1].(map[string]dbus.Variant)
			if castOk2 == false {
				fmt.Println("failed cast to map[string]dbus.Variant: ", sig)
				continue
			}

			switch intf {
			case "org.freedesktop.UPower.Device":
				pType, err := GetDeviceProperty(sig.Path, "org.freedesktop.UPower.Device."+Type)
				if err != nil {
					fmt.Println("failed to get UPower.Device Type", sig)
					continue
				}
				uiType := pType.Value().(uint32)

				switch uiType {
				case Battery:

					nativePath, err := GetDeviceProperty(sig.Path, "org.freedesktop.UPower.Device."+NativePath)
					if err != nil {
						fmt.Println("failed to get UPower.Device nativePath", sig)
						continue
					}
					nativePathStr := nativePath.Value().(string)

					properties := make([]string, 0)
					for key, val := range changedProperties {
						switch key {
						case Percentage:
							perc := val.Value().(float64)
							str := "percentage=" + strconv.FormatFloat(perc, 'f', -1, 64)
							properties = append(properties, str)
						case EnergyRate:
							erate := val.Value().(float64)
							str := "rate=" + strconv.FormatFloat(erate, 'f', -1, 64)
							properties = append(properties, str)
						case TimeToEmpty:
							tte := time.Duration(time.Duration(val.Value().(int64)) * time.Second)
							str := "time_to_empty=" + tte.String()
							properties = append(properties, str)
						case TimeToFull:
							ttf := time.Duration(time.Duration(val.Value().(int64)) * time.Second)
							str := "time_to_full=" + ttf.String()
							properties = append(properties, str)
						case State:
							st := val.Value().(uint32)
							str := "state=" + strconv.Itoa(int(st))
							properties = append(properties, str)
							// case Energy:
							// 	energ := val.Value().(float64)
							// 	fmt.Println("--- energy (W):", energ)
						}
					}
					if len(properties) > 0 {
						datalog.Append("data;" + nativePathStr + ";" + strings.Join(properties, ","))
					}
				case LinePower:
					for key, val := range changedProperties {
						switch key {
						case Online:
							online := val.Value().(bool)
							if online == true {
								datalog.Append("online")
							} else {
								datalog.Append("offline")
							}
						}
					}
				}
			}

		case "org.freedesktop.login1.Manager.PrepareForSleep":
			prepare, castOk := sig.Body[0].(bool)
			if castOk == false {
				fmt.Println("failed cast to bool: ", sig)
				continue
			}
			if prepare == true {
				datalog.Append("sleep")
				datalog.AddSuspendEvent()
			} else {
				datalog.Append("resume")
			}
		}
	}
	return nil
}
