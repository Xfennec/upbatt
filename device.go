package main

import (
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
