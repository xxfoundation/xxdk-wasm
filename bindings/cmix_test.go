////////////////////////////////////////////////////////////////////////////////
// Copyright © 2020 xx network SEZC                                           //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file                                                               //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm
// +build js,wasm

package bindings

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
