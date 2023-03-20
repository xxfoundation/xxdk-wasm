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

// Tests that the map representing DMClient returned by newDMClientJS contains
// all of the methods on DMClient.
func Test_newDMClientJS(t *testing.T) {
	dmcType := reflect.TypeOf(&DMClient{})

	dmc := newDMClientJS(&bindings.DMClient{})
	if len(dmc) != dmcType.NumMethod() {
		t.Errorf("DMClient JS object does not have all methods."+
			"\nexpected: %d\nreceived: %d", dmcType.NumMethod(), len(dmc))
	}

	for i := 0; i < dmcType.NumMethod(); i++ {
		method := dmcType.Method(i)

		if _, exists := dmc[method.Name]; !exists {
			t.Errorf("Method %s does not exist.", method.Name)
		}
	}
}

// Tests that DMClient has all the methods that
// [bindings.DMClient] has.
func Test_DMClientMethods(t *testing.T) {
	dmcType := reflect.TypeOf(&DMClient{})
	binDmcType := reflect.TypeOf(&bindings.DMClient{})

	var numOfExcludedFields int
	if _, exists := dmcType.MethodByName("GetDatabaseName"); !exists {
		t.Errorf("GetDatabaseName was not found.")
	} else {
		numOfExcludedFields++
	}

	nm := dmcType.NumMethod() - numOfExcludedFields
	if binDmcType.NumMethod() != nm {
		t.Errorf("WASM DMClient object does not have all methods from "+
			"bindings.\nexpected: %d\nreceived: %d", binDmcType.NumMethod(), nm)
	}

	for i := 0; i < binDmcType.NumMethod(); i++ {
		method := binDmcType.Method(i)

		if _, exists := dmcType.MethodByName(method.Name); !exists {
			t.Errorf("Method %s does not exist.", method.Name)
		}
	}
}

// Tests that the map representing DMDbCipher returned by newDMDbCipherJS
// contains all of the methods on DMDbCipher.
func Test_newDMDbCipherJS(t *testing.T) {
	cipherType := reflect.TypeOf(&DMDbCipher{})

	cipher := newDMDbCipherJS(&bindings.DMDbCipher{})
	if len(cipher) != cipherType.NumMethod() {
		t.Errorf("DMDbCipher JS object does not have all methods."+
			"\nexpected: %d\nreceived: %d", cipherType.NumMethod(), len(cipher))
	}

	for i := 0; i < cipherType.NumMethod(); i++ {
		method := cipherType.Method(i)

		if _, exists := cipher[method.Name]; !exists {
			t.Errorf("Method %s does not exist.", method.Name)
		}
	}
}

// Tests that DMDbCipher has all the methods that [bindings.DMDbCipher] has.
func Test_DMDbCipherMethods(t *testing.T) {
	cipherType := reflect.TypeOf(&DMDbCipher{})
	binCipherType := reflect.TypeOf(&bindings.DMDbCipher{})

	if binCipherType.NumMethod() != cipherType.NumMethod() {
		t.Errorf("WASM DMDbCipher object does not have all methods from "+
			"bindings.\nexpected: %d\nreceived: %d",
			binCipherType.NumMethod(), cipherType.NumMethod())
	}

	for i := 0; i < binCipherType.NumMethod(); i++ {
		method := binCipherType.Method(i)

		if _, exists := cipherType.MethodByName(method.Name); !exists {
			t.Errorf("Method %s does not exist.", method.Name)
		}
	}
}
