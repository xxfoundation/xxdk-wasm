////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm
// +build js,wasm

package wasm

import (
	"gitlab.com/elixxir/client/bindings"
	"reflect"
	"testing"
)

// Tests that the map representing Cmix returned by newCmixJS contains all of
// the methods on Cmix.
func Test_newCmixJS(t *testing.T) {
	cmixType := reflect.TypeOf(&Cmix{})

	cmix := newCmixJS(&bindings.Cmix{})
	if len(cmix) != cmixType.NumMethod() {
		t.Errorf("Cmix JS object does not have all methods."+
			"\nexpected: %d\nreceived: %d", cmixType.NumMethod(), len(cmix))
	}

	for i := 0; i < cmixType.NumMethod(); i++ {
		method := cmixType.Method(i)

		if _, exists := cmix[method.Name]; !exists {
			t.Errorf("Method %s does not exist.", method.Name)
		}
	}
}

// Tests that Cmix has all the methods that [bindings.Cmix] has.
func Test_CmixMethods(t *testing.T) {
	cmixType := reflect.TypeOf(&Cmix{})
	binCmixType := reflect.TypeOf(&bindings.Cmix{})

	if binCmixType.NumMethod() != cmixType.NumMethod() {
		t.Errorf("WASM Cmix object does not have all methods from bindings."+
			"\nexpected: %d\nreceived: %d",
			binCmixType.NumMethod(), cmixType.NumMethod())
	}

	for i := 0; i < binCmixType.NumMethod(); i++ {
		method := binCmixType.Method(i)

		if _, exists := cmixType.MethodByName(method.Name); !exists {
			t.Errorf("Method %s does not exist.", method.Name)
		}
	}
}
