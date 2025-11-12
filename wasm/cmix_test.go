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

// Tests that Cmix has all the methods that [bindings.Cmix] has, except for
// methods that are intentionally excluded from WASM.
func Test_CmixMethods(t *testing.T) {
	cmixType := reflect.TypeOf(&Cmix{})
	binCmixType := reflect.TypeOf(&bindings.Cmix{})

	// Methods excluded from WASM wrapper
	excludedMethods := map[string]bool{
		"GetRemoteKV": true, // Removed - was only used by synchronized Cmix which doesn't work
	}

	expectedMethods := binCmixType.NumMethod() - len(excludedMethods)
	if expectedMethods != cmixType.NumMethod() {
		t.Errorf("WASM Cmix object does not have expected number of methods."+
			"\nexpected: %d (bindings: %d - excluded: %d)\nreceived: %d",
			expectedMethods, binCmixType.NumMethod(), len(excludedMethods), cmixType.NumMethod())
	}

	for i := 0; i < binCmixType.NumMethod(); i++ {
		method := binCmixType.Method(i)

		// Skip excluded methods
		if excludedMethods[method.Name] {
			continue
		}

		if _, exists := cmixType.MethodByName(method.Name); !exists {
			t.Errorf("Method %s does not exist.", method.Name)
		}
	}
}
