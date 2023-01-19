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

// Tests that FileTransfer has all the methods that [bindings.FileTransfer] has.
func Test_FileTransferMethods(t *testing.T) {
	ftType := reflect.TypeOf(&FileTransfer{})
	binFtType := reflect.TypeOf(&bindings.FileTransfer{})

	if binFtType.NumMethod() != ftType.NumMethod() {
		t.Errorf("WASM FileTransfer object does not have all methods from "+
			"bindings.\nexpected: %d\nreceived: %d",
			binFtType.NumMethod(), ftType.NumMethod())
	}

	for i := 0; i < binFtType.NumMethod(); i++ {
		method := binFtType.Method(i)

		if _, exists := ftType.MethodByName(method.Name); !exists {
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

// Tests that FilePartTracker has all the methods that
// [bindings.FilePartTracker] has.
func Test_FilePartTrackerMethods(t *testing.T) {
	fptType := reflect.TypeOf(&FilePartTracker{})
	binFptType := reflect.TypeOf(&bindings.FilePartTracker{})

	if binFptType.NumMethod() != fptType.NumMethod() {
		t.Errorf("WASM FilePartTracker object does not have all methods from "+
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
