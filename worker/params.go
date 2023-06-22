////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package worker

import "time"

// Params are parameters used in the [MessageManager].
type Params struct {
	// MessageLogging indicates if a DEBUG message should be printed every time
	// a message is sent or received.
	MessageLogging bool

	// ResponseTimeout is the default timeout to wait for a response before
	// timing out and returning an error.
	ResponseTimeout time.Duration
}

// DefaultParams returns the default parameters.
func DefaultParams() Params {
	return Params{
		MessageLogging:  false,
		ResponseTimeout: 30 * time.Second,
	}
}
