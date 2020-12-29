// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package driverreg_test

import (
	"fmt"
	"log"

	"periph.io/x/conn/driver/driverreg"
)

func Example() {
	// Make sure periph is initialized.
	// TODO: Use host.Init(). It is not used in this example to prevent circular
	// go package import.
	state, err := driverreg.Init()
	if err != nil {
		log.Fatal(err)
	}

	// Prints the loaded driver.
	fmt.Printf("Using drivers:\n")
	for _, driver := range state.Loaded {
		fmt.Printf("- %s\n", driver)
	}

	// Prints the driver that were skipped as irrelevant on the platform.
	fmt.Printf("Drivers skipped:\n")
	for _, failure := range state.Skipped {
		fmt.Printf("- %s: %s\n", failure.D, failure.Err)
	}

	// Having drivers failing to load may not require process termination. It
	// is possible to continue to run in partial failure mode.
	fmt.Printf("Drivers failed to load:\n")
	for _, failure := range state.Failed {
		fmt.Printf("- %s: %v\n", failure.D, failure.Err)
	}

	// Use pins, buses, devices, etc.
}
