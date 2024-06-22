package main

import (
	"math"
	"sync/atomic"
	"time"

	"gobot.io/x/gobot/v2/platforms/joystick"
)

type pair struct {
	x float64
	y float64
}

var (
	joystickAdaptor *joystick.Adaptor
	stick           *joystick.Driver

	leftX, leftY, rightX, rightY atomic.Value
)

const offset = 32767.0

func startJoystick() {
	joystickAdaptor = joystick.NewAdaptor("0")
	stick = joystick.NewDriver(joystickAdaptor, joystick.Dualshock3)

	joystickAdaptor.Connect()
	stick.Start()

	leftX.Store(float64(0.0))
	leftY.Store(float64(0.0))
	rightX.Store(float64(0.0))
	rightY.Store(float64(0.0))

	stick.On(joystick.TrianglePress, func(data interface{}) {
		drone.TakeOff()
	})

	stick.On(joystick.XPress, func(data interface{}) {
		drone.Land()
	})

	stick.On(joystick.LeftX, func(data interface{}) {
		val := float64(data.(int))
		leftX.Store(val)
	})

	stick.On(joystick.LeftY, func(data interface{}) {
		val := float64(data.(int))
		leftY.Store(val)
	})

	stick.On(joystick.RightX, func(data interface{}) {
		val := float64(data.(int))
		rightX.Store(val)
	})

	stick.On(joystick.RightY, func(data interface{}) {
		val := float64(data.(int))
		rightY.Store(val)
	})

	go handleLeftJoystick()
	go handleRightJoystick()
}

func handleRightJoystick() {
	for {
		rightStick := getRightStick()

		switch {
		case rightStick.y < -10:
			drone.Forward(ValidatePitch(rightStick.y, offset))
		case rightStick.y > 10:
			drone.Backward(ValidatePitch(rightStick.y, offset))
		default:
			drone.Forward(0)
		}

		switch {
		case rightStick.x > 10:
			drone.Right(ValidatePitch(rightStick.x, offset))
		case rightStick.x < -10:
			drone.Left(ValidatePitch(rightStick.x, offset))
		default:
			drone.Right(0)
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func handleLeftJoystick() {
	for {
		leftStick := getLeftStick()

		switch {
		case leftStick.y < -10:
			drone.Up(ValidatePitch(leftStick.y, offset))
		case leftStick.y > 10:
			drone.Down(ValidatePitch(leftStick.y, offset))
		default:
			drone.Up(0)
		}

		switch {
		case leftStick.x > 20:
			drone.Clockwise(ValidatePitch(leftStick.x, offset))
		case leftStick.x < -20:
			drone.CounterClockwise(ValidatePitch(leftStick.x, offset))
		default:
			drone.Clockwise(0)
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func getLeftStick() pair {
	s := pair{x: 0, y: 0}
	s.x = leftX.Load().(float64)
	s.y = leftY.Load().(float64)
	return s
}

func getRightStick() pair {
	s := pair{x: 0, y: 0}
	s.x = rightX.Load().(float64)
	s.y = rightY.Load().(float64)
	return s
}

// ValidatePitch helps validate pitch values such as those created by
// a joystick to values between 0-100 that are required as
// params to Parrot Minidrone PCMDs
func ValidatePitch(data float64, offset float64) int {
	value := math.Abs(data) / offset
	if value >= 0.1 {
		if value <= 1.0 {
			return int((float64(int(value*100)) / 100) * 100)
		}
		return 100
	}
	return 0
}
