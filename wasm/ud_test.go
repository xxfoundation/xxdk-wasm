////////////////////////////////////////////////////////////////////////////////
// Copyright © 2020 xx network SEZC                                           //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file                                                               //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"gitlab.com/elixxir/client/bindings"
	"reflect"
	"testing"
)

// Tests that the map representing UserDiscovery returned by newUserDiscoveryJS
// contains all of the methods on UserDiscovery.
func Test_newUserDiscoveryJS(t *testing.T) {
	udType := reflect.TypeOf(&UserDiscovery{})

	ud := newUserDiscoveryJS(&bindings.UserDiscovery{})
	if len(ud) != udType.NumMethod() {
		t.Errorf("UserDiscovery JS object does not have all methods."+
			"\nexpected: %d\nreceived: %d", udType.NumMethod(), len(ud))
	}

	for i := 0; i < udType.NumMethod(); i++ {
		method := udType.Method(i)

		if _, exists := ud[method.Name]; !exists {
			t.Errorf("Method %s does not exist.", method.Name)
		}
	}
}
