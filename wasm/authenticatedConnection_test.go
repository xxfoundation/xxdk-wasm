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

// Tests that AuthenticatedConnection has all the methods that
// [bindings.AuthenticatedConnection] has.
func Test_AuthenticatedConnectionMethods(t *testing.T) {
	authType := reflect.TypeOf(&AuthenticatedConnection{})
	binAuthType := reflect.TypeOf(&bindings.AuthenticatedConnection{})

	if binAuthType.NumMethod() != authType.NumMethod() {
		t.Errorf("WASM AuthenticatedConnection object does not have all "+
			"methods from bindings.\nexpected: %d\nreceived: %d",
			binAuthType.NumMethod(), authType.NumMethod())
	}

	for i := 0; i < binAuthType.NumMethod(); i++ {
		method := binAuthType.Method(i)

		if _, exists := authType.MethodByName(method.Name); !exists {
			t.Errorf("Method %s does not exist.", method.Name)
		}
	}
}
