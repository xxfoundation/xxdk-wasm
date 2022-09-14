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

// Tests that the map representing Channel returned by newChannelJS contains
// all of the methods on Channel.
func Test_newChannelJS(t *testing.T) {
	chanType := reflect.TypeOf(&Channel{})

	ch := newChannelJS(&bindings.Channel{})
	if len(ch) != chanType.NumMethod() {
		t.Errorf("Channel JS object does not have all methods."+
			"\nexpected: %d\nreceived: %d", chanType.NumMethod(), len(ch))
	}

	for i := 0; i < chanType.NumMethod(); i++ {
		method := chanType.Method(i)

		if _, exists := ch[method.Name]; !exists {
			t.Errorf("Method %s does not exist.", method.Name)
		}
	}
}

// Tests that Channel has all the methods that [bindings.Channel] has.
func Test_ChannelMethods(t *testing.T) {
	chanType := reflect.TypeOf(&Channel{})
	binChanType := reflect.TypeOf(&bindings.Channel{})

	if binChanType.NumMethod() != chanType.NumMethod() {
		t.Errorf("WASM Channel object does not have all methods from bindings."+
			"\nexpected: %d\nreceived: %d",
			binChanType.NumMethod(), chanType.NumMethod())
	}

	for i := 0; i < binChanType.NumMethod(); i++ {
		method := binChanType.Method(i)

		if _, exists := chanType.MethodByName(method.Name); !exists {
			t.Errorf("Method %s does not exist.", method.Name)
		}
	}
}
