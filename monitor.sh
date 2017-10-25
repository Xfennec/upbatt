#!/bin/bash

# dump all signals from UPower (for debug)
dbus-monitor --system "type='signal',sender='org.freedesktop.UPower'"
