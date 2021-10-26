// Copyright 2021 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package gpioutil

import (
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpiotest"
)

func TestPulseIn_Success(t *testing.T) {
	edgesChan := make(chan gpio.Level, 1)
	clock := clockwork.NewFakeClock()

	pin := gpiotest.Pin{
		EdgesChan: edgesChan,
		L:         gpio.Low,
		Clock:     clock,
	}
	// insert for pin.In emptying buffer
	edgesChan <- gpio.High

	go func() {
		for len(edgesChan) != 0 {
		}
		edgesChan <- gpio.High

		// insert for pin.In emptying buffer
		edgesChan <- gpio.High
		for len(edgesChan) != 0 {
		}

		clock.Advance(time.Second)
		edgesChan <- gpio.Low
	}()

	duration, err := pulseInWithClock(&pin, gpio.High, -1, clock)
	if err != nil {
		t.Fatal("shouldn't have any error")
	}

	if duration != time.Second {
		t.Fatal("it should takes 1 second")
	}
}

func TestPulseIn_Timeout_1(t *testing.T) {
	done := make(chan struct{})

	edgesChan := make(chan gpio.Level)
	clock := clockwork.NewFakeClock()

	pin := gpiotest.Pin{
		EdgesChan: edgesChan,
		L:         gpio.Low,
		Clock:     clock,
	}

	go func() {
		for {
			select {
			case <-done:
				return
			default:
				clock.Advance(time.Second)
			}
		}
	}()

	duration, err := pulseInWithClock(&pin, gpio.High, time.Second, clock)
	if err != nil {
		t.Fatal("shouldn't have any error")
	}

	close(done)

	if duration != 0 {
		t.Fatal("it should returns 0 for timeout")
	}
}

func TestPulseIn_Timeout_2(t *testing.T) {
	done := make(chan struct{})

	edgesChan := make(chan gpio.Level, 1)
	clock := clockwork.NewFakeClock()

	pin := gpiotest.Pin{
		EdgesChan: edgesChan,
		L:         gpio.Low,
		Clock:     clock,
	}
	// insert for pin.In emptying buffer
	edgesChan <- gpio.High

	go func() {
		for len(edgesChan) != 0 {
		}
		edgesChan <- gpio.High

		// insert for pin.In emptying buffer
		edgesChan <- gpio.High
		for len(edgesChan) != 0 {
		}

		for {
			select {
			case <-done:
				return
			default:
				clock.Advance(time.Second)
			}
		}
	}()

	duration, err := pulseInWithClock(&pin, gpio.High, time.Second, clock)
	if err != nil {
		t.Fatal("shouldn't have any error")
	}

	close(done)

	if duration != 0 {
		t.Fatal("it should returns 0 for timeout")
	}
}
