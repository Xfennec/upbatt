package main

import (
	"errors"
	"fmt"

	"github.com/godbus/dbus"
)

// Unknown shared by multiple properties
const Unknown = 0

// Possible power sources
const (
	LinePower = 1
	Battery   = 2
	UPS       = 3
	Monitor   = 4
	Mouse     = 5
	Keyboard  = 6
	PDA       = 7
	Phone     = 8
)

// Possible states
const (
	Charging         = 1
	Discharging      = 2
	Empty            = 3
	FullyCharged     = 4
	PendingCharge    = 5
	PendingDischarge = 6
)

// Possible technologys
const (
	LithiumIon           = 1
	LithiumPolymer       = 2
	LithiumIronPhosphate = 3
	LeadAcid             = 4
	NickelCadmium        = 5
	NickelMetalHydride   = 6
)

// Properties
const (
	NativePath       = "NativePath"
	Vendor           = "Vendor"
	Model            = "Model"
	Serial           = "Serial"
	UpdateTime       = "UpdateTime"
	Type             = "Type"
	PowerSupply      = "PowerSupply"
	HasHistory       = "HasHistory"
	HasStatistics    = "HasStatistics"
	Online           = "Online"
	Energy           = "Energy"
	EnergyEmpty      = "EnergyEmpty"
	EnergyFull       = "EnergyFull"
	EnergyFullDesign = "EnergyFullDesign"
	EnergyRate       = "EnergyRate"
	Voltage          = "Voltage"
	TimeToEmpty      = "TimeToEmpty"
	TimeToFull       = "TimeToFull"
	Percentage       = "Percentage"
	IsPresent        = "IsPresent"
	State            = "State"
	IsRechargeable   = "IsRechargeable"
	Capacity         = "Capacity"
	Technology       = "Technology"
	RecallNotice     = "RecallNotice"
	RecallVendor     = "RecallVendor"
	RecallURL        = "RecallUrl"
)

// Stats stores some basic battery informations
// (see BatteryStats below)
type Stats struct {
	Percentage float64
	State      int
}

// BatteryStats store some basic informations about a battery from
// current battery state and log.
// It's needed for log reconciliation on server startup.
type BatteryStats struct {
	Path       dbus.ObjectPath
	NativePath string
	Real       Stats
	Log        Stats
}

// EnumerateDevices enumerate all power objects on the system.
func EnumerateDevices() (devices []dbus.ObjectPath, err error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return
	}
	obj := conn.Object("org.freedesktop.UPower", "/org/freedesktop/UPower")

	call := obj.Call("org.freedesktop.UPower.EnumerateDevices", 0)
	if call.Err != nil {
		return nil, call.Err
	}

	if err := call.Store(&devices); err != nil {
		return nil, err
	}

	return
}

// GetDeviceProperty Get UPower device property
func GetDeviceProperty(dev dbus.ObjectPath, p string) (v dbus.Variant, err error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return
	}
	obj := conn.Object("org.freedesktop.UPower", dev)

	v, err = obj.GetProperty(p)
	if err != nil {
		return
	}

	return
}

// GetRealPowerState returns the current online/offline state
func GetRealPowerState(devices []dbus.ObjectPath) (string, error) {
	// get current online/offline state
	var powerReal = ""
	for _, device := range devices {
		pType, err := GetDeviceProperty(device, "org.freedesktop.UPower.Device."+Type)
		if err != nil {
			fmt.Println("failed to get UPower.Device Type", device)
			continue
		}
		uiType := pType.Value().(uint32)
		if uiType == LinePower {
			if powerReal != "" {
				return "", errors.New("multiple LinePower devices are not supported yet (report this case, please)")
			}
			pOnline, err := GetDeviceProperty(device, "org.freedesktop.UPower.Device."+Online)
			if err != nil {
				fmt.Println("failed to get UPower.Device Online state", device)
				continue
			}
			bOnline := pOnline.Value().(bool)
			if bOnline == true {
				powerReal = online
			} else {
				powerReal = offline
			}
		}
	}
	if powerReal == "" {
		return "", errors.New("can't get actual power status (online/offline)")
	}

	return powerReal, nil
}

// GetRealBatteriesStats returns a list of BatteryStats for all batteries
func GetRealBatteriesStats(devices []dbus.ObjectPath) []BatteryStats {
	batteries := make([]BatteryStats, 0)

	for _, device := range devices {
		pType, err := GetDeviceProperty(device, "org.freedesktop.UPower.Device."+Type)
		if err != nil {
			fmt.Println("failed to get UPower.Device Type", device)
			continue
		}
		uiType := pType.Value().(uint32)
		if uiType == Battery {
			var battery BatteryStats
			battery.Path = device
			battery.Real.Percentage = -1
			battery.Real.State = -1
			battery.Log.Percentage = -1
			battery.Log.State = -1

			nativePath, err := GetDeviceProperty(device, "org.freedesktop.UPower.Device."+NativePath)
			if err != nil {
				fmt.Println("failed to get UPower.Device NativePath", device)
				continue
			}
			battery.NativePath = nativePath.Value().(string)

			// get state & percentage
			pPercentage, err := GetDeviceProperty(device, "org.freedesktop.UPower.Device."+Percentage)
			if err != nil {
				fmt.Println("failed to get UPower.Device Percentage", device)
			} else {
				battery.Real.Percentage = pPercentage.Value().(float64)
			}

			pState, err := GetDeviceProperty(device, "org.freedesktop.UPower.Device."+State)
			if err != nil {
				fmt.Println("failed to get UPower.Device State", device)
			} else {
				battery.Real.State = int(pState.Value().(uint32))
			}

			batteries = append(batteries, battery)
		}
	}
	return batteries
}
