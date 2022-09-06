////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"gitlab.com/elixxir/client/bindings"
	"reflect"
	"testing"
)

// Tests that the map representing E2e returned by newE2eJS contains all of the
// methods on E2e.
func Test_newE2eJS(t *testing.T) {
	e2eType := reflect.TypeOf(&E2e{})

	e2e := newE2eJS(&bindings.E2e{})
	if len(e2e) != e2eType.NumMethod() {
		t.Errorf("E2e JS object does not have all methods."+
			"\nexpected: %d\nreceived: %d", e2eType.NumMethod(), len(e2e))
	}

	for i := 0; i < e2eType.NumMethod(); i++ {
		method := e2eType.Method(i)

		if _, exists := e2e[method.Name]; !exists {
			t.Errorf("Method %s does not exist.", method.Name)
		}
	}
}
