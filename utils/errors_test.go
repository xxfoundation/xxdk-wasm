////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package utils

import (
	"fmt"
	"github.com/pkg/errors"
	"testing"
)

// Tests that TestJsError returns a Javascript Error object with the expected
// message.
func TestJsError(t *testing.T) {
	err := errors.New("test error")
	expectedErr := err.Error()
	jsError := JsError(err).Get("message").String()

	if jsError != expectedErr {
		t.Errorf("Failed to get expected error message."+
			"\nexpected: %s\nreceived: %s", expectedErr, jsError)
	}
}

// Tests that TestJsTrace returns a Javascript Error object with the expected
// message and stack trace.
func TestJsTrace(t *testing.T) {
	err := errors.New("test error")
	expectedErr := fmt.Sprintf("%+v", err)
	jsError := JsTrace(err).Get("message").String()

	if jsError != expectedErr {
		t.Errorf("Failed to get expected error message."+
			"\nexpected: %s\nreceived: %s", expectedErr, jsError)
	}
}
