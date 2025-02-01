// Copyright 2025 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package gpio

import (
	"errors"
	"time"

	"periph.io/x/conn/v3"
	"periph.io/x/conn/v3/pin"
)

type GPIOValue uint64

// Implementations that don't implement specific interface methods should
// return ErrGroupFeatureNotImplemented as the error to allow clients to
// generically check for the condition.
var ErrGroupFeatureNotImplemented = errors.New("gpio group feature not implemented")

// Group is an interface that an IO device can implement to manipulate multiple
// IO Pins at one time. Performing GPIO Operations in this manner can dramatically
// simplify code and reduce IO Operation latency.
//
// Device specific code can also provide methods that return a Group that operates
// on a subset of the pins the device supports.
type Group interface {
	// The set of GPIO pins that make up this group. Implementations will
	// typically Use gpio.PinIO, gpio.PinIn, or gpio.PinOut as the actual return
	// value type.
	Pins() []pin.Pin
	// Given a specific pin offset within the group, return that pin.
	// For example, a pin group may be GPIO pins 3,5,7,9 in that order.
	// ByOffset(1) returns GPIO pin 5.
	ByOffset(offset int) pin.Pin
	// Given the unique name of a GPIO pin, return that pin.
	ByName(name string) pin.Pin
	// Given the specific GPIO pin number, return the corresponding
	// pin from the group.
	ByNumber(number int) pin.Pin
	// Out writes the specified bitwise value to the pins. Bit 0 corresponds to
	// the first pin in the set, bit 1 the second, etc. Only pins within the
	// group that have mask bit set are modified. For example, if you have 8
	// pins within the group and you want to write the value 0x0a to the lower
	// 4 pins, you would use a mask of 0x0f.
	//
	// If the device doesn't support write operations, implementations should
	// return gpio.ErrGroupFeatureNotImplemented.
	Out(value, mask GPIOValue) error
	// Read reads the pins within the group, and returns the  value, ANDed with
	// mask. If the device doesn't support read operations, implementations
	// should return gpio.ErrGroupFeatureNotImplemented.
	Read(mask GPIOValue) (GPIOValue, error)
	// WaitForEdge blocks for a GPIO line change event to happen. If the does
	// not implement gpio.PinIn, or doesn't support this capability,
	// implementations should return gpio.ErrGroupFeatureNotImplemented.
	//
	// Number is the GPIO pin number within the group that had an edge change.
	WaitForEdge(timeout time.Duration) (number int, edge Edge, err error)
	// conn.Resource brings in resource.Halt(), and fmt.Stringer
	conn.Resource
}
