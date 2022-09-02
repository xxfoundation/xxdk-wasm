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
