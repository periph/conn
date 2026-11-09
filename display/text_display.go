// Copyright 2024 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Character-display interfaces for cursors, backlights, contrast, and text output.

package display

import (
	"errors"
)

// CursorDirection specifies movement relative to the current cursor.
type CursorDirection int

const (
	// Backward moves the cursor one unit back.
	Backward CursorDirection = iota
	// Forward moves the cursor one unit forward.
	Forward
	// Up moves the cursor up one row.
	Up
	// Down moves the cursor down one row.
	Down
)

// CursorMode describes a visible cursor style.
type CursorMode int

const (
	// CursorOff hides the cursor.
	CursorOff CursorMode = iota
	// CursorUnderline displays an underline cursor.
	CursorUnderline
	// CursorBlock displays a block cursor.
	CursorBlock
	// CursorBlink enables cursor blinking.
	CursorBlink
)

// TextDisplay represents an interface to a basic character device. It provides
// standard methods implemented by a majority of character LCD devices. Pixel
// type displays can implement this interface if desired.
type TextDisplay interface {
	// Enable/Disable auto scroll
	AutoScroll(enabled bool) (err error)
	// Return the number of columns the display supports
	Cols() int
	// Clear the display and move the cursor home.
	Clear() (err error)
	// Set the cursor mode. You can pass multiple arguments.
	// Cursor(CursorOff, CursorUnderline)
	//
	// Implementations should return an error if the value of mode is not
	// mode>= CursorOff && mode <= CursorBlink
	Cursor(mode ...CursorMode) (err error)
	// Move the cursor home (MinRow(),MinCol())
	Home() (err error)
	// Return the min column position.
	MinCol() int
	// Return the min row position.
	MinRow() int
	// Move the cursor forward or backward.
	Move(dir CursorDirection) (err error)
	// Move the cursor to arbitrary position. Implementations should return an
	// error if row < MinRow() || row > (Rows()-MinRow()), or col < MinCol()
	// || col > (Cols()-MinCol())
	MoveTo(row, col int) (err error)
	// Return the number of rows the display supports.
	Rows() int
	// Turn the display on / off
	Display(on bool) (err error)
	// return info about the display.
	String() string
	// Write a set of bytes to the display.
	Write(p []byte) (n int, err error)
	// Write a string output to the display.
	WriteString(text string) (n int, err error)
}

// Intensity represents a backlight intensity.
type Intensity int

// DisplayBacklight is implemented by displays with a monochrome backlight.  Displays that
// support RGB Backlights should implement this as well for maximum
// compatibility.
//
// Many units that support this command write the value to EEPROM, which has a
// finite number of writes. To turn the unit on/off, use TextDisplay.Display().
type DisplayBacklight interface {
	Backlight(intensity Intensity) error
}

// DisplayRGBBacklight is implemented by displays with an RGB backlight.
type DisplayRGBBacklight interface {
	RGBBacklight(red, green, blue Intensity) error
}

// Contrast represents a display contrast level.
type Contrast int

// DisplayContrast is implemented by displays with programmable contrast.
// As with SetBacklight(), many devices serialize the value to EEPROM,
// which support only a finite number of writes, so this should be used
// sparingly.
type DisplayContrast interface {
	Contrast(contrast Contrast) error
}

// ErrNotImplemented indicates an unsupported display operation.
var ErrNotImplemented = errors.New("not implemented")

// ErrInvalidCommand indicates an invalid display command.
var ErrInvalidCommand = errors.New("invalid command")
