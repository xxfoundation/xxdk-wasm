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

// Tests that the map representing GroupChat returned by newGroupChatJS contains
// all of the methods on GroupChat.
func Test_newGroupChatJS(t *testing.T) {
	gcType := reflect.TypeOf(&GroupChat{})

	gc := newGroupChatJS(&bindings.GroupChat{})
	if len(gc) != gcType.NumMethod() {
		t.Errorf("GroupChat JS object does not have all methods."+
			"\nexpected: %d\nreceived: %d", gcType.NumMethod(), len(gc))
	}

	for i := 0; i < gcType.NumMethod(); i++ {
		method := gcType.Method(i)

		if _, exists := gc[method.Name]; !exists {
			t.Errorf("Method %s does not exist.", method.Name)
		}
	}
}

// Tests that the map representing Group returned by newGroupJS contains all of
// the methods on Group.
func Test_newGroupJS(t *testing.T) {
	gType := reflect.TypeOf(&Group{})

	g := newGroupJS(&bindings.Group{})
	if len(g) != gType.NumMethod() {
		t.Errorf("Group JS object does not have all methods."+
			"\nexpected: %d\nreceived: %d", gType.NumMethod(), len(g))
	}

	for i := 0; i < gType.NumMethod(); i++ {
		method := gType.Method(i)

		if _, exists := g[method.Name]; !exists {
			t.Errorf("Method %s does not exist.", method.Name)
		}
	}
}
