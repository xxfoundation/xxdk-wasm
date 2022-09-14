////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
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

// Tests that Backup has all the methods that [bindings.Backup] has.
func Test_BackupMethods(t *testing.T) {
	backupType := reflect.TypeOf(&Backup{})
	binBackupType := reflect.TypeOf(&bindings.Backup{})

	if binBackupType.NumMethod() != backupType.NumMethod() {
		t.Errorf("WASM Backup object does not have all methods from bindings."+
			"\nexpected: %d\nreceived: %d",
			binBackupType.NumMethod(), backupType.NumMethod())
	}

	for i := 0; i < binBackupType.NumMethod(); i++ {
		method := binBackupType.Method(i)

		if _, exists := backupType.MethodByName(method.Name); !exists {
			t.Errorf("Method %s does not exist.", method.Name)
		}
	}
}
