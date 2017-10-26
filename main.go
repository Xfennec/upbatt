package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/godbus/dbus"
)

func signalPump(ch chan *dbus.Signal, datalog *DataLog) error {

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
				pType, err := GetProperty(sig.Path, "org.freedesktop.UPower.Device."+Type)
				if err != nil {
					fmt.Println("failed to get UPower.Device Type", sig)
					continue
				}
				uiType := pType.Value().(uint32)

				switch uiType {
				case Battery:
					// fmt.Println("--- PropertiesChanged on UPower.Device for a battery:")
					// TimeToEmpty / TimeToFull
					// EnergyRate (meaning is different for Discharging and Charging I guess…)
					// Percentage
					// (Energy [60W])
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
						// case Energy:
						// 	energ := val.Value().(float64)
						// 	fmt.Println("--- energy (W):", energ)
						default:
							// fmt.Println(key, val)
						}
					}
					if len(properties) > 0 {
						datalog.Append("data;" + strings.Join(properties, ","))
					}
				case LinePower:
					// fmt.Println("--- PropertiesChanged on UPower.Device for a powerline:")
					for key, val := range changedProperties {
						switch key {
						case Online:
							online := val.Value().(bool)
							if online == true {
								datalog.Append("online")
							} else {
								datalog.Append("offline")
							}
						default:
							// fmt.Println(key, val)
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
			} else {
				datalog.Append("resume")
			}
		}
	}
	return nil
}

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

	if err := AliveSchedule(10*time.Second, datalog); err != nil {
		return err
	}

	fmt.Println("ready.")
	if err := signalPump(ch, datalog); err != nil {
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
		fmt.Println("we're the client.")
	}
}
