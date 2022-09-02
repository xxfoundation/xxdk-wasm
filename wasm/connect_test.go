////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package wasm

import (
	"gitlab.com/elixxir/client/bindings"
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
