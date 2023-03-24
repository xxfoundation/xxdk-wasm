////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package utils

import (
	"fmt"
	"syscall/js"
)

// JsError converts the error to a Javascript Error.
func JsError(err error) js.Value {
	return Error.New(err.Error())
}

// JsTrace converts the error to a Javascript Error that includes the error's
// stack trace.
func JsTrace(err error) js.Value {
	return Error.New(fmt.Sprintf("%+v", err))
}

// Throw function stub to throws Javascript exceptions. The exception must be
// one of the defined Exception below. Any other error types will result in an
// error.
func Throw(exception Exception, err error) {
	throw(exception, fmt.Sprintf("%+v", err))
}

func throw(_ Exception, message string) {
	panic(message)
}

// Exception are the possible Javascript error types that can be thrown.
type Exception string

const (
	// EvalError occurs when error has occurred in the eval() function.
	//
	// Deprecated: This exception is not thrown by JavaScript anymore, however
	// the EvalError object remains for compatibility.
	EvalError Exception = "EvalError"

	// RangeError occurs when a numeric variable or parameter is outside its
	// valid range.
	RangeError Exception = "RangeError"

	// ReferenceError occurs when a variable that does not exist (or hasn't yet
	// been initialized) in the current scope is referenced.
	ReferenceError Exception = "ReferenceError"

	// SyntaxError occurs when trying to interpret syntactically invalid code.
	SyntaxError Exception = "SyntaxError"

	// TypeError occurs when an operation could not be performed, typically (but
	// not exclusively) when a value is not of the expected type.
	//
	// A TypeError may be thrown when:
	//  - an operand or argument passed to a function is incompatible with the
	//    type expected by that operator or function; or
	//  - when attempting to modify a value that cannot be changed; or
	//  - when attempting to use a value in an inappropriate way.
	TypeError Exception = "TypeError"

	// URIError occurs when a global URI handling function was used in a wrong
	// way.
	URIError Exception = "URIError"
)
