// Copyright 2021 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package gpioutil

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpiotest"
)

func TestPulseIn_Success(t *testing.T) {
	edgesChan := make(chan gpio.Level)
	clock := clockwork.NewFakeClock()

	pin := pulseInPin{
		sleeps: []time.Duration{0, time.Second},
		Pin: gpiotest.Pin{
			EdgesChan: edgesChan,
			L:         gpio.Low,
			Clock:     clock,
		},
	}

	go func() {
		edgesChan <- gpio.High

		// There is no timeout on PulseIn so there isn't any After in WaitForEdges.
		// Here we simulate one in In function.
		clock.BlockUntil(1)
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
	edgesChan := make(chan gpio.Level)
	clock := clockwork.NewFakeClock()

	pin := pulseInPin{
		Pin: gpiotest.Pin{
			EdgesChan: edgesChan,
			L:         gpio.Low,
			Clock:     clock,
		},
	}

	go func() {
		clock.BlockUntil(1)
		clock.Advance(time.Second)
	}()

	duration, err := pulseInWithClock(&pin, gpio.High, time.Second, clock)
	if err != nil {
		t.Fatal("shouldn't have any error")
	}

	if duration != 0 {
		t.Fatal("it should returns 0 for timeout")
	}
}

func TestPulseIn_Timeout_2(t *testing.T) {
	edgesChan := make(chan gpio.Level)
	clock := clockwork.NewFakeClock()

	pin := pulseInPin{
		Pin: gpiotest.Pin{
			EdgesChan: edgesChan,
			L:         gpio.Low,
			Clock:     clock,
		},
	}

	go func() {
		edgesChan <- gpio.High

		// there is a call for after in the first WaitForEdge and there is another one in the second WaitForEdge.
		clock.BlockUntil(2)

		clock.Advance(time.Second)
	}()

	duration, err := pulseInWithClock(&pin, gpio.High, time.Second, clock)
	if err != nil {
		t.Fatal("shouldn't have any error")
	}
	if duration != 0 {
		t.Fatal("it should returns 0 for timeout")
	}
}

type pulseInPin struct {
	gpiotest.Pin

	sleeps []time.Duration
}

func (p *pulseInPin) In(pull gpio.Pull, edge gpio.Edge) error {
	p.Lock()
	defer p.Unlock()
	p.P = pull
	if pull == gpio.PullDown {
		p.L = gpio.Low
	} else if pull == gpio.PullUp {
		p.L = gpio.High
	}

	if edge != gpio.NoEdge && p.EdgesChan == nil {
		return errors.New("gpiotest: please set p.EdgesChan first")
	}

	if len(p.sleeps) > 0 {
		if p.sleeps[0] != 0 {
			p.Clock.Sleep(p.sleeps[0])
		}
		p.sleeps = p.sleeps[1:]
	}

	fmt.Printf("there is a %s\n", edge)

	return nil
}
