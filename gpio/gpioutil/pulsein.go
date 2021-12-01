// Copyright 2021 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package gpioutil

import (
	"time"

	"github.com/jonboulle/clockwork"
	"periph.io/x/conn/v3/gpio"
)

// PulseIn reads a pulse (either HIGH or LOW) on a pin.
//
// For example, if lvl is HIGH, PulseIn() waits for the pin to go from LOW to HIGH, starts timing,
// then waits for the pin to go LOW and stops timing.
//
// Returns the length of the pulse as a time.Duration or gives up and returns 0
// if no complete pulse was received within the timeout.
func PulseIn(pin gpio.PinIn, lvl gpio.Level, t time.Duration) (time.Duration, error) {
	return pulseInWithClock(pin, lvl, t, clockwork.NewRealClock())
}

func pulseInWithClock(pin gpio.PinIn, lvl gpio.Level, t time.Duration, clock clockwork.Clock) (time.Duration, error) {
	e1 := gpio.FallingEdge
	e2 := gpio.RisingEdge

	if lvl == gpio.High {
		e1 = gpio.RisingEdge
		e2 = gpio.FallingEdge
	}

	if err := pin.In(gpio.PullNoChange, e1); err != nil {
		return 0, err
	}

	if !pin.WaitForEdge(t) {
		return 0, nil
	}

	now := clock.Now()

	if err := pin.In(gpio.PullNoChange, e2); err != nil {
		return 0, err
	}

	if !pin.WaitForEdge(t) {
		return 0, nil
	}

	return clock.Since(now), nil
}
