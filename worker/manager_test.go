////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package worker

import (
	"syscall/js"
	"testing"
)

// Tests that newWorkerOptions returns a Javascript object with the expected
// type, credentials, and name fields.
func Test_newWorkerOptions(t *testing.T) {
	for _, workerType := range []string{"classic", "module"} {
		for _, credentials := range []string{"omit", "same-origin", "include"} {
			for _, name := range []string{"name1", "name2", "name3"} {
				opts := newWorkerOptions(workerType, credentials, name)

				optsJS := js.ValueOf(opts)

				typeJS := optsJS.Get("type").String()
				if typeJS != workerType {
					t.Errorf("Unexected type (type:%s credentials:%s name:%s)"+
						"\nexpected: %s\nreceived: %s",
						workerType, credentials, name, workerType, typeJS)
				}

				credentialsJS := optsJS.Get("credentials").String()
				if typeJS != workerType {
					t.Errorf("Unexected credentials (type:%s credentials:%s "+
						"name:%s)\nexpected: %s\nreceived: %s", workerType,
						credentials, name, credentials, credentialsJS)
				}

				nameJS := optsJS.Get("name").String()
				if typeJS != workerType {
					t.Errorf("Unexected name (type:%s credentials:%s name:%s)"+
						"\nexpected: %s\nreceived: %s",
						workerType, credentials, name, name, nameJS)
				}
			}
		}
	}
}
