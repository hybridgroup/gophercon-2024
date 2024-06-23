package main

import (
	"os"
	"time"

	powerup "github.com/hybridgroup/tinygo-powerup"
	"tinygo.org/x/bluetooth"
)

var deviceAddress string

var (
	adapter = bluetooth.DefaultAdapter
	device  bluetooth.Device
	ch      = make(chan bluetooth.ScanResult, 1)

	airplane *powerup.Airplane
)

func main() {
	deviceAddress = connectAddress()

	println("enabling...")
	must("enable BLE interface", adapter.Enable())

	println("start scan...")
	must("start scan", adapter.Scan(scanHandler))

	var err error
	select {
	case result := <-ch:
		device, err = adapter.Connect(result.Address, bluetooth.ConnectionParams{})
		must("connect to peripheral device", err)

		println("connected to ", result.Address.String())
	}

	defer device.Disconnect()

	airplane = powerup.NewAirplane(&device)
	err = airplane.Start()
	if err != nil {
		println(err)
	}

	time.Sleep(3 * time.Second)

	startJoystick()

	defer airplane.Stop()

	select {}
}

func scanHandler(a *bluetooth.Adapter, d bluetooth.ScanResult) {
	println("device:", d.Address.String(), d.RSSI, d.LocalName())
	if d.Address.String() == deviceAddress {
		a.StopScan()
		ch <- d
	}
}

func must(action string, err error) {
	if err != nil {
		for {
			println("failed to " + action + ": " + err.Error())
			time.Sleep(time.Second)
		}
	}
}

func connectAddress() string {
	if len(os.Args) < 2 {
		println("you must pass the Bluetooth address of the PowerUp you want to connect to as the first argument")
		os.Exit(1)
	}

	address := os.Args[1]

	return address
}
