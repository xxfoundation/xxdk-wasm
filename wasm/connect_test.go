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

// Tests that the map representing Connection returned by newConnectJS contains
// all of the methods on Connection.
func Test_newConnectJS(t *testing.T) {
	connType := reflect.TypeOf(&Connection{})

	conn := newConnectJS(&bindings.Connection{})
	if len(conn) != connType.NumMethod() {
		t.Errorf("Connection JS object does not have all methods."+
			"\nexpected: %d\nreceived: %d", connType.NumMethod(), len(conn))
	}

	for i := 0; i < connType.NumMethod(); i++ {
		method := connType.Method(i)

		if _, exists := conn[method.Name]; !exists {
			t.Errorf("Method %s does not exist.", method.Name)
		}
	}
}

// Tests that Connection has all the methods that [bindings.Connection] has.
func Test_ConnectionMethods(t *testing.T) {
	connType := reflect.TypeOf(&Connection{})
	binConnType := reflect.TypeOf(&bindings.Connection{})

	if binConnType.NumMethod() != connType.NumMethod() {
		t.Errorf("WASM Connection object does not have all methods from "+
			"bindings.\nexpected: %d\nreceived: %d",
			binConnType.NumMethod(), connType.NumMethod())
	}

	for i := 0; i < binConnType.NumMethod(); i++ {
		method := binConnType.Method(i)

		if _, exists := connType.MethodByName(method.Name); !exists {
			t.Errorf("Method %s does not exist.", method.Name)
		}
	}
}
