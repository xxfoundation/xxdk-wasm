////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package worker

import (
	"testing"
)

// Tests NewManager
func TestNewManager(t *testing.T) {
	aURL := "/index.js"
	aName := "Name"
	m, err := NewManager(aURL, aName)
	if err != nil {
		t.Fatalf("Failed to create new Manager: %+v", err)
	}

	t.Logf("Manager: %+v", m)
}

// Tests Manager.SendMessage
func TestManager_SendMessage(t *testing.T) {
	// TODO
}

// Tests Manager.receiveMessage
func TestManager_receiveMessage(t *testing.T) {
	// TODO
}

// Tests Manager.getCallback
func TestManager_getCallback(t *testing.T) {
	// TODO
}

// Tests Manager.RegisterCallback
func TestManager_RegisterCallback(t *testing.T) {
	// TODO
}

// Tests Manager.registerReplyCallback
func TestManager_registerReplyCallback(t *testing.T) {
	// TODO
}

// Tests that Manager.getNextID returns the expected ID for various Tags.
func TestManager_getNextID(t *testing.T) {
	m := &Manager{
		callbacks:   make(map[Tag]map[uint64]ReceptionCallback),
		responseIDs: make(map[Tag]uint64),
	}

	for _, tag := range []Tag{readyTag, "test", "A", "B", "C"} {
		id := m.getNextID(tag)
		if id != initID {
			t.Errorf("ID for new tag %q is not initID."+
				"\nexpected: %d\nreceived: %d", tag, initID, id)
		}

		for j := uint64(1); j < 100; j++ {
			id = m.getNextID(tag)
			if id != j {
				t.Errorf("Unexpected ID for tag %q."+
					"\nexpected: %d\nreceived: %d", tag, j, id)
			}
		}
	}
}

// Tests Manager.GetWorker
func TestManager_GetWorker(t *testing.T) {
	// TODO
}

////////////////////////////////////////////////////////////////////////////////
// Javascript Call Wrappers                                                   //
////////////////////////////////////////////////////////////////////////////////

// Tests Manager.addEventListeners
func TestManager_addEventListeners(t *testing.T) {
	// TODO
}

// Tests Manager.postMessage
func TestManager_postMessage(t *testing.T) {
	// TODO
}

// Tests Manager.Terminate
func TestManager_Terminate(t *testing.T) {
	// TODO
}

// Tests that newWorkerOptions returns a Javascript object with the expected
// type, credentials, and name fields.
func Test_newWorkerOptions(t *testing.T) {
	for i, workerType := range []string{"classic", "module"} {
		for j, credentials := range []string{"omit", "same-origin", "include"} {
			for k, name := range []string{"name1", "name2", "name3"} {
				opts := newWorkerOptions(workerType, credentials, name)

				v := opts.Get("type").String()
				if v != workerType {
					t.Errorf("Unexpected type (%d, %d, %d)."+
						"\nexpected: %s\nreceived: %s", i, j, k, workerType, v)
				}

				v = opts.Get("credentials").String()
				if v != credentials {
					t.Errorf("Unexpected credentials (%d, %d, %d)."+
						"\nexpected: %s\nreceived: %s", i, j, k, credentials, v)
				}

				v = opts.Get("name").String()
				if v != name {
					t.Errorf("Unexpected name (%d, %d, %d)."+
						"\nexpected: %s\nreceived: %s", i, j, k, name, v)
				}
			}
		}
	}
}
