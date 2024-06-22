package main

import (
	"os"
	"time"

	minidrone "github.com/hybridgroup/tinygo-minidrone"
	"tinygo.org/x/bluetooth"
)

var deviceAddress = connectAddress()

var (
	adapter = bluetooth.DefaultAdapter
	device  bluetooth.Device
	ch      = make(chan bluetooth.ScanResult, 1)

	drone *minidrone.Minidrone
)

func main() {
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

	drone = minidrone.NewMinidrone(&device)
	drone.PilotingStateChange(func(state, substate int) {
		switch state {
		case minidrone.PilotingStateFlyingStateChanged:
			println("FlightStateChange", minidrone.FlyingState(substate))
		default:
			println("PilotingStateChange", state, substate)
		}
	})

	err = drone.Start()
	if err != nil {
		println(err)
	}

	time.Sleep(3 * time.Second)

	err = drone.TakeOff()
	if err != nil {
		println(err)
	}

	done := make(chan bool)
	go flightPlan(done)

	select {
	case <-done:
	case <-time.After(30 * time.Second):
	}

	drone.Land()
	time.Sleep(time.Second * 3)

	drone.Halt()
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
		println("you must pass the Bluetooth address of the minidrone y0u want to connect to as the first argument")
		os.Exit(1)
	}

	address := os.Args[1]

	return address
}

func flightPlan(done chan bool) {
	drone.Hover()
	time.Sleep(time.Second * 5)

	drone.Up(20)
	time.Sleep(time.Second * 3)

	drone.Hover()
	time.Sleep(time.Second * 5)

	drone.FrontFlip()
	time.Sleep(time.Second * 10)

	drone.Hover()
	time.Sleep(time.Second * 5)

	done <- true
}
