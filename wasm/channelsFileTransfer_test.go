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

// Tests that the map representing ChannelsFileTransfer returned by
// newChannelsFileTransferJS contains all of the methods on ChannelsFileTransfer.
func Test_newChannelsFileTransferJS(t *testing.T) {
	cftType := reflect.TypeOf(&ChannelsFileTransfer{})

	ft := newChannelsFileTransferJS(&bindings.ChannelsFileTransfer{})
	if len(ft) != cftType.NumMethod() {
		t.Errorf("ChannelsFileTransfer JS object does not have all methods."+
			"\nexpected: %d\nreceived: %d", cftType.NumMethod(), len(ft))
	}

	for i := 0; i < cftType.NumMethod(); i++ {
		method := cftType.Method(i)

		if _, exists := ft[method.Name]; !exists {
			t.Errorf("Method %s does not exist.", method.Name)
		}
	}
}

// Tests that ChannelsFileTransfer has all the methods that
// [bindings.ChannelsFileTransfer] has.
func Test_ChannelsFileTransferMethods(t *testing.T) {
	cftType := reflect.TypeOf(&ChannelsFileTransfer{})
	binCftType := reflect.TypeOf(&bindings.ChannelsFileTransfer{})

	if binCftType.NumMethod() != cftType.NumMethod() {
		t.Errorf("WASM ChannelsFileTransfer object does not have all methods "+
			"from bindings.\nexpected: %d\nreceived: %d",
			binCftType.NumMethod(), cftType.NumMethod())
	}

	for i := 0; i < binCftType.NumMethod(); i++ {
		method := binCftType.Method(i)

		if _, exists := cftType.MethodByName(method.Name); !exists {
			t.Errorf("Method %s does not exist.", method.Name)
		}
	}
}

// Tests that the map representing ChFilePartTracker returned by
// newChFilePartTrackerJS contains all of the methods on ChFilePartTracker.
func Test_newChFilePartTrackerJS(t *testing.T) {
	fptType := reflect.TypeOf(&FilePartTracker{})

	fpt := newChFilePartTrackerJS(&bindings.ChFilePartTracker{})
	if len(fpt) != fptType.NumMethod() {
		t.Errorf("ChFilePartTracker JS object does not have all methods."+
			"\nexpected: %d\nreceived: %d", fptType.NumMethod(), len(fpt))
	}

	for i := 0; i < fptType.NumMethod(); i++ {
		method := fptType.Method(i)

		if _, exists := fpt[method.Name]; !exists {
			t.Errorf("Method %s does not exist.", method.Name)
		}
	}
}

// Tests that ChFilePartTracker has all the methods that
// [bindings.ChFilePartTracker] has.
func Test_ChFilePartTrackerMethods(t *testing.T) {
	fptType := reflect.TypeOf(&ChFilePartTracker{})
	binFptType := reflect.TypeOf(&bindings.ChFilePartTracker{})

	if binFptType.NumMethod() != fptType.NumMethod() {
		t.Errorf("WASM ChFilePartTracker object does not have all methods from "+
			"bindings.\nexpected: %d\nreceived: %d",
			binFptType.NumMethod(), fptType.NumMethod())
	}

	for i := 0; i < binFptType.NumMethod(); i++ {
		method := binFptType.Method(i)

		if _, exists := fptType.MethodByName(method.Name); !exists {
			t.Errorf("Method %s does not exist.", method.Name)
		}
	}
}
