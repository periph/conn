package gpioutil

import (
	"time"

	"periph.io/x/conn/v3/gpio"
)

// PluseIn reads a pulse (either HIGH or LOW) on a pin.
// For example, if lvl is HIGH, PulseIn() waits for the pin to go from LOW to HIGH, starts timing,
// then waits for the pin to go LOW and stops timing.
// Returns the length of the pulse as a time.Duration or gives up and returns 0
// if no complete pulse was received within the timeout.
func PulseIn(pin gpio.PinIn, lvl gpio.Level, t time.Duration) (time.Duration, error) {
	var e1, e2 gpio.Edge

	if lvl == gpio.High {
		e1 = gpio.RisingEdge
		e2 = gpio.FallingEdge
	} else {
		e1 = gpio.FallingEdge
		e2 = gpio.RisingEdge
	}

	if err := pin.In(gpio.PullNoChange, e1); err != nil {
		return 0, err
	}

	if !pin.WaitForEdge(t) {
		return 0, nil
	}

	now := time.Now()

	if err := pin.In(gpio.PullNoChange, e2); err != nil {
		return 0, err
	}

	if !pin.WaitForEdge(t) {
		return 0, nil
	}

	return time.Since(now), nil
}
