////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"reflect"
	"testing"
)

// Tests that the map representing DbCipher returned by
// newDbCipherJS contains all the methods on DbCipher.
func Test_newChannelDbCipherJS(t *testing.T) {
	cipherType := reflect.TypeOf(&DbCipher{})

	cipher := newDbCipherJS(&DbCipher{})
	if len(cipher) != cipherType.NumMethod() {
		t.Errorf("DbCipher JS object does not have all methods."+
			"\nexpected: %d\nreceived: %d", cipherType.NumMethod(), len(cipher))
	}

	for i := 0; i < cipherType.NumMethod(); i++ {
		method := cipherType.Method(i)

		if _, exists := cipher[method.Name]; !exists {
			t.Errorf("Method %s does not exist.", method.Name)
		}
	}
}
