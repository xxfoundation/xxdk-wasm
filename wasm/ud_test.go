////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"gitlab.com/elixxir/client/v4/bindings"
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

// Tests that UserDiscovery has all the methods that [bindings.UserDiscovery]
// has.
func Test_UserDiscoveryMethods(t *testing.T) {
	udType := reflect.TypeOf(&UserDiscovery{})
	binUdType := reflect.TypeOf(&bindings.UserDiscovery{})

	if binUdType.NumMethod() != udType.NumMethod() {
		t.Errorf("WASM UserDiscovery object does not have all methods from "+
			"bindings.\nexpected: %d\nreceived: %d",
			binUdType.NumMethod(), udType.NumMethod())
	}

	for i := 0; i < binUdType.NumMethod(); i++ {
		method := binUdType.Method(i)

		if _, exists := udType.MethodByName(method.Name); !exists {
			t.Errorf("Method %s does not exist.", method.Name)
		}
	}
}
