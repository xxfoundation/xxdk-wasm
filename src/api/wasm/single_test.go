////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"gitlab.com/elixxir/client/v4/bindings"
	"reflect"
	"testing"
)

// Tests that the map representing Stopper returned by newStopperJS contains all
// of the methods on Stopper.
func Test_newStopperJS(t *testing.T) {
	stopperType := reflect.TypeOf(&Stopper{})

	s := newStopperJS(&stopper{})
	if len(s) != stopperType.NumMethod() {
		t.Errorf("Stopper JS object does not have all methods."+
			"\nexpected: %d\nreceived: %d", stopperType.NumMethod(), len(s))
	}

	for i := 0; i < stopperType.NumMethod(); i++ {
		method := stopperType.Method(i)

		if _, exists := s[method.Name]; !exists {
			t.Errorf("Method %s does not exist.", method.Name)
		}
	}
}

// Tests that Stopper has all the methods that [bindings.Stopper] has.
func Test_StopperMethods(t *testing.T) {
	stopperType := reflect.TypeOf(&Stopper{})
	binStopperType := reflect.TypeOf(bindings.Stopper(&stopper{}))

	if binStopperType.NumMethod() != stopperType.NumMethod() {
		t.Errorf("WASM Stopper object does not have all methods from bindings."+
			"\nexpected: %d\nreceived: %d",
			binStopperType.NumMethod(), stopperType.NumMethod())
	}

	for i := 0; i < binStopperType.NumMethod(); i++ {
		method := binStopperType.Method(i)

		if _, exists := stopperType.MethodByName(method.Name); !exists {
			t.Errorf("Method %s does not exist.", method.Name)
		}
	}
}

type stopper struct{}

func (s *stopper) Stop() {}
