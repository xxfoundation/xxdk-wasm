////////////////////////////////////////////////////////////////////////////////
// Copyright © 2020 xx network SEZC                                           //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file                                                               //
////////////////////////////////////////////////////////////////////////////////

package wasm

import (
	"gitlab.com/elixxir/client/bindings"
	"reflect"
	"testing"
)

// Tests that the map representing AuthenticatedConnection returned by
// newAuthenticatedConnectionJS contains all of the methods on
// AuthenticatedConnection.
func Test_newAuthenticatedConnectionJS(t *testing.T) {
	acType := reflect.TypeOf(&AuthenticatedConnection{})

	ch := newAuthenticatedConnectionJS(&bindings.AuthenticatedConnection{})
	if len(ch) != acType.NumMethod() {
		t.Errorf("AuthenticatedConnection JS object does not have all methods."+
			"\nexpected: %d\nreceived: %d", acType.NumMethod(), len(ch))
	}

	for i := 0; i < acType.NumMethod(); i++ {
		method := acType.Method(i)

		if _, exists := ch[method.Name]; !exists {
			t.Errorf("Method %s does not exist.", method.Name)
		}
	}
}
