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

// Tests that GroupChat has all the methods that [bindings.GroupChat] has.
func Test_GroupChatMethods(t *testing.T) {
	gcType := reflect.TypeOf(&GroupChat{})
	binGcType := reflect.TypeOf(&bindings.GroupChat{})

	if binGcType.NumMethod() != gcType.NumMethod() {
		t.Errorf("WASM GroupChat object does not have all methods from "+
			"bindings.\nexpected: %d\nreceived: %d",
			binGcType.NumMethod(), gcType.NumMethod())
	}

	for i := 0; i < binGcType.NumMethod(); i++ {
		method := binGcType.Method(i)

		if _, exists := gcType.MethodByName(method.Name); !exists {
			t.Errorf("Method %s does not exist.", method.Name)
		}
	}
}

// Tests that the map representing Group returned by newGroupJS contains all of
// the methods on Group.
func Test_newGroupJS(t *testing.T) {
	grpType := reflect.TypeOf(&Group{})

	g := newGroupJS(&bindings.Group{})
	if len(g) != grpType.NumMethod() {
		t.Errorf("Group JS object does not have all methods."+
			"\nexpected: %d\nreceived: %d", grpType.NumMethod(), len(g))
	}

	for i := 0; i < grpType.NumMethod(); i++ {
		method := grpType.Method(i)

		if _, exists := g[method.Name]; !exists {
			t.Errorf("Method %s does not exist.", method.Name)
		}
	}
}

// Tests that Group has all the methods that [bindings.Group] has.
func Test_GroupMethods(t *testing.T) {
	grpType := reflect.TypeOf(&Group{})
	binGrpType := reflect.TypeOf(&bindings.Group{})

	if binGrpType.NumMethod() != grpType.NumMethod() {
		t.Errorf("WASM Group object does not have all methods from bindings."+
			"\nexpected: %d\nreceived: %d",
			binGrpType.NumMethod(), grpType.NumMethod())
	}

	for i := 0; i < binGrpType.NumMethod(); i++ {
		method := binGrpType.Method(i)

		if _, exists := grpType.MethodByName(method.Name); !exists {
			t.Errorf("Method %s does not exist.", method.Name)
		}
	}
}
