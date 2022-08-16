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

// Tests that the map representing Backup returned by newBackupJS contains all
// of the methods on Backup.
func Test_newBackupJS(t *testing.T) {
	buType := reflect.TypeOf(&Backup{})

	b := newBackupJS(&bindings.Backup{})
	if len(b) != buType.NumMethod() {
		t.Errorf("Backup JS object does not have all methods."+
			"\nexpected: %d\nreceived: %d", buType.NumMethod(), len(b))
	}

	for i := 0; i < buType.NumMethod(); i++ {
		method := buType.Method(i)

		if _, exists := b[method.Name]; !exists {
			t.Errorf("Method %s does not exist.", method.Name)
		}
	}
}
