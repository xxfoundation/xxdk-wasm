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

// Tests that the map representing ChannelsManager returned by
// newChannelsManagerJS contains all of the methods on ChannelsManager.
func Test_newChannelsManagerJS(t *testing.T) {
	cmType := reflect.TypeOf(&ChannelsManager{})

	e2e := newChannelsManagerJS(&bindings.ChannelsManager{})
	if len(e2e) != cmType.NumMethod() {
		t.Errorf("ChannelsManager JS object does not have all methods."+
			"\nexpected: %d\nreceived: %d", cmType.NumMethod(), len(e2e))
	}

	for i := 0; i < cmType.NumMethod(); i++ {
		method := cmType.Method(i)

		if _, exists := e2e[method.Name]; !exists {
			t.Errorf("Method %s does not exist.", method.Name)
		}
	}
}

// Tests that ChannelsManager has all the methods that
// [bindings.ChannelsManager] has.
func Test_ChannelsManagerMethods(t *testing.T) {
	cmType := reflect.TypeOf(&ChannelsManager{})
	binCmType := reflect.TypeOf(&bindings.ChannelsManager{})

	if binCmType.NumMethod() != cmType.NumMethod() {
		t.Errorf("WASM ChannelsManager object does not have all methods from "+
			"bindings.\nexpected: %d\nreceived: %d",
			binCmType.NumMethod(), cmType.NumMethod())
	}

	for i := 0; i < binCmType.NumMethod(); i++ {
		method := binCmType.Method(i)

		if _, exists := cmType.MethodByName(method.Name); !exists {
			t.Errorf("Method %s does not exist.", method.Name)
		}
	}
}
