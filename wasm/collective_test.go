////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"reflect"
	"testing"

	"gitlab.com/elixxir/client/v4/bindings"
)

// Tests that the map representing RemoteKV returned by newRemoteKvJS contains
// all of the methods on RemoteKV.
func Test_newRemoteKvJS(t *testing.T) {
	rkvType := reflect.TypeOf(&RemoteKV{})

	rkv := newRemoteKvJS(&bindings.RemoteKV{})
	if len(rkv) != rkvType.NumMethod() {
		t.Errorf("RemoteKV JS object does not have all methods."+
			"\nexpected: %d\nreceived: %d", rkvType.NumMethod(), len(rkv))
	}

	for i := 0; i < rkvType.NumMethod(); i++ {
		method := rkvType.Method(i)

		if _, exists := rkv[method.Name]; !exists {
			t.Errorf("Method %s does not exist.", method.Name)
		}
	}
}

// Tests that RemoteKV has all the methods that [bindings.RemoteKV] has.
func Test_RemoteKVMethods(t *testing.T) {
	rkvType := reflect.TypeOf(&RemoteKV{})
	binRkvType := reflect.TypeOf(&bindings.RemoteKV{})

	if binRkvType.NumMethod() != rkvType.NumMethod() {
		t.Errorf("WASM RemoteKV object does not have all methods from "+
			"bindings.\nexpected: %d\nreceived: %d",
			binRkvType.NumMethod(), rkvType.NumMethod())
	}

	for i := 0; i < binRkvType.NumMethod(); i++ {
		method := binRkvType.Method(i)

		if _, exists := rkvType.MethodByName(method.Name); !exists {
			t.Errorf("Method %s does not exist.", method.Name)
		}
	}
}
