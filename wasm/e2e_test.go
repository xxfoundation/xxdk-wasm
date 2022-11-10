////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"gitlab.com/elixxir/client/v5/bindings"
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

// Tests that E2e has all the methods that [bindings.E2e] has.
func Test_E2eMethods(t *testing.T) {
	e2eType := reflect.TypeOf(&E2e{})
	binE2eType := reflect.TypeOf(&bindings.E2e{})

	if binE2eType.NumMethod() != e2eType.NumMethod() {
		t.Errorf("WASM E2e object does not have all methods from bindings."+
			"\nexpected: %d\nreceived: %d",
			binE2eType.NumMethod(), e2eType.NumMethod())
	}

	for i := 0; i < binE2eType.NumMethod(); i++ {
		method := binE2eType.Method(i)

		if _, exists := e2eType.MethodByName(method.Name); !exists {
			t.Errorf("Method %s does not exist.", method.Name)
		}
	}
}
