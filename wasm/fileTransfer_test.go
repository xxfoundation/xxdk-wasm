////////////////////////////////////////////////////////////////////////////////
// Copyright © 2020 xx network SEZC                                           //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file                                                               //
////////////////////////////////////////////////////////////////////////////////

package wasm

import (
	"gitlab.com/elixxir/client/bindings"
	"reflect"
	"testing"
)

// Tests that the map representing FileTransfer returned by newFileTransferJS
// contains all of the methods on FileTransfer.
func Test_newFileTransferJS(t *testing.T) {
	ftType := reflect.TypeOf(&FileTransfer{})

	ft := newFileTransferJS(&bindings.FileTransfer{})
	if len(ft) != ftType.NumMethod() {
		t.Errorf("File Transfer JS object does not have all methods."+
			"\nexpected: %d\nreceived: %d", ftType.NumMethod(), len(ft))
	}

	for i := 0; i < ftType.NumMethod(); i++ {
		method := ftType.Method(i)

		if _, exists := ft[method.Name]; !exists {
			t.Errorf("Method %s does not exist.", method.Name)
		}
	}
}

// Tests that the map representing FilePartTracker returned by
// newFilePartTrackerJS contains all of the methods on FilePartTracker.
func Test_newFilePartTrackerJS(t *testing.T) {
	fptType := reflect.TypeOf(&FilePartTracker{})

	fpt := newFilePartTrackerJS(&bindings.FilePartTracker{})
	if len(fpt) != fptType.NumMethod() {
		t.Errorf("File part tracker JS object does not have all methods."+
			"\nexpected: %d\nreceived: %d", fptType.NumMethod(), len(fpt))
	}

	for i := 0; i < fptType.NumMethod(); i++ {
		method := fptType.Method(i)

		if _, exists := fpt[method.Name]; !exists {
			t.Errorf("Method %s does not exist.", method.Name)
		}
	}
}
