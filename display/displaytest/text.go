// Copyright 2024 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package displaytest

import (
	"errors"
	"fmt"
	"time"

	"periph.io/x/conn/v3/display"
)

// TestTextDisplay exercises the methods provided by the interface. It can be
// used interactively as a quick smoke test of an implementation, and from test
// routines. This doesn't test brightness or contrast to avoid EEPROM wear
// issues.
func TestTextDisplay(dev display.TextDisplay, interactive bool) []error {
	result := make([]error, 0)
	var err error

	pauseTime := time.Millisecond
	if interactive {
		pauseTime = 3 * time.Second
	}
	// Turn the dev on and write the String() value.
	if err = dev.Display(true); err != nil {
		result = append(result, err)
	}

	if err = dev.Clear(); err != nil {
		result = append(result, err)
	}

	if _, err = dev.WriteString(dev.String()); err != nil {
		result = append(result, err)
	}
	time.Sleep(pauseTime)

	if err = dev.Clear(); err != nil {
		result = append(result, err)
	}
	_, err = dev.WriteString("Auto Scroll Test")
	if err != nil {
		result = append(result, err)
	}
	time.Sleep(pauseTime)
	err = dev.AutoScroll(true)
	if err != nil {
		result = append(result, err)
	}
	// Test Display fill
	for line := range dev.Rows() {
		c := rune('A')
		if err = dev.MoveTo(dev.MinRow()+line, dev.MinCol()); err != nil {
			result = append(result, err)
		}
		for col := range dev.Cols() {
			if col%5 == 0 && col > 0 {
				_, err = dev.Write([]byte{byte(' ')})
			} else {
				_, err = dev.Write([]byte{byte(c)})
			}
			if err != nil {
				result = append(result, err)
			}
			c = c + 1
		}
	}
	// Test AutoScroll working
	time.Sleep(pauseTime)
	nWritten, err := dev.WriteString("auto scroll happen")
	if err != nil {
		result = append(result, err)
	}
	if nWritten != 18 {
		result = append(result, fmt.Errorf("dev.WriteString() expected %d chars written, received %d", 18, nWritten))
	}
	time.Sleep(pauseTime)
	if err = dev.AutoScroll(false); err != nil {
		result = append(result, err)
	}
	time.Sleep(pauseTime)

	// Test Absolute Positioning
	if err = dev.Clear(); err != nil {
		result = append(result, err)
	}
	if _, err = dev.WriteString("Absolute Positioning"); err != nil {
		result = append(result, err)
	}
	time.Sleep(pauseTime)
	if err = dev.Clear(); err != nil {
		result = append(result, err)
	}
	for ix := range dev.Rows() {
		if err = dev.MoveTo(dev.MinRow()+ix, dev.MinCol()+ix); err != nil {
			result = append(result, err)
		}
		if _, err = dev.WriteString(fmt.Sprintf("(%d,%d)", dev.MinRow()+ix, dev.MinCol()+ix)); err != nil {
			result = append(result, err)
		}
	}
	time.Sleep(pauseTime)

	// Test that MoveTo returns error for invalid coordinates
	moveCases := []struct {
		row int
		col int
	}{
		{row: dev.MinRow() - 1, col: dev.MinCol()},
		{row: dev.MinRow(), col: dev.MinCol() - 1},
		{row: dev.Rows() + 1, col: dev.Cols()},
		{row: dev.Rows(), col: dev.Cols() + 1},
	}
	for _, tc := range moveCases {
		if err = dev.MoveTo(tc.row, tc.col); err == nil {
			result = append(result, fmt.Errorf("did not receive expected error on MoveTo(%d,%d)", tc.row, tc.col))
		}
	}

	// Test Cursor Modes
	if err = dev.Clear(); err != nil {
		result = append(result, err)
	}
	modes := []string{"Off", "Underline", "Block", "Blink"}
	for ix := display.CursorOff; ix <= display.CursorBlink; ix++ {
		if err = dev.MoveTo(dev.MinRow()/2+1, dev.MinCol()); err != nil {
			result = append(result, err)
		}
		if _, err = dev.WriteString("Cursor: " + modes[ix]); err != nil {
			result = append(result, err)
		}
		if err = dev.Cursor(ix); err != nil {
			result = append(result, err)
		}
		time.Sleep(pauseTime)
		if err = dev.Cursor(display.CursorOff); err != nil {
			result = append(result, err)
		}
		if err = dev.Clear(); err != nil {
			result = append(result, err)
		}
	}
	if err = dev.Cursor(display.CursorBlink + 1); err == nil {
		result = append(result, errors.New("did not receive expected error on Cursor() with invalid value"))
	}

	// Test Move Forward and Backward. 2 Should overwrite the 1
	if err = dev.Clear(); err != nil {
		result = append(result, err)
	}
	if _, err = dev.WriteString("Testing >"); err != nil {
		result = append(result, err)
	}
	if err = dev.Move(display.Forward); err != nil {
		result = append(result, err)
	}
	if err = dev.Move(display.Forward); err != nil {
		result = append(result, err)
	}
	for ix := range 10 {
		if _, err := dev.WriteString(fmt.Sprintf("%d", ix)); err != nil {
			result = append(result, err)
		}
		time.Sleep(pauseTime)
		if err := dev.Move(display.Backward); err != nil {
			result = append(result, err)
		}
	}
	if err = dev.Move(display.Down + 1); err == nil {
		result = append(result, errors.New("did not receive expected error on Move() with invalid value"))
	}

	// Test Display on/off
	if err = dev.Clear(); err != nil {
		result = append(result, err)
	}
	if _, err = dev.WriteString("Set dev off"); err != nil {
		result = append(result, err)
	}
	time.Sleep(pauseTime)
	if err = dev.Display(false); err != nil {
		result = append(result, err)
	}
	time.Sleep(pauseTime)
	if err = dev.Display(true); err != nil {
		result = append(result, err)
	}
	if err = dev.Clear(); err != nil {
		result = append(result, err)
	}
	if _, err = dev.WriteString("Set dev on"); err != nil {
		result = append(result, err)
	}
	time.Sleep(pauseTime)

	if interactive {
		for _, e := range result {
			if !errors.Is(err, display.ErrNotImplemented) {
				fmt.Println(e)
			}
		}
	}
	return result
}
