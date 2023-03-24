////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package worker

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

// Tests Manager.receiveMessage calls the expected callback.
func TestManager_receiveMessage(t *testing.T) {
	m := &Manager{callbacks: make(map[Tag]map[uint64]ReceptionCallback)}

	msg := Message{Tag: readyTag, ID: 5}
	cbChan := make(chan struct{})
	cb := func([]byte) { cbChan <- struct{}{} }
	m.callbacks[msg.Tag] = map[uint64]ReceptionCallback{msg.ID: cb}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to JSON marshal Message: %+v", err)
	}

	go func() {
		select {
		case <-cbChan:
		case <-time.After(10 * time.Millisecond):
			t.Error("Timed out waiting for callback to be called.")
		}
	}()

	err = m.receiveMessage(data)
	if err != nil {
		t.Errorf("Failed to receive message: %+v", err)
	}
}

// Tests Manager.getCallback returns the expected callback and deletes only the
// given callback when deleteCB is true.
func TestManager_getCallback(t *testing.T) {
	m := &Manager{callbacks: make(map[Tag]map[uint64]ReceptionCallback)}

	// Add new callback and check that it is returned by getCallback
	tag, id1 := readyTag, uint64(5)
	cb := func([]byte) {}
	m.callbacks[tag] = map[uint64]ReceptionCallback{id1: cb}

	received, err := m.getCallback(tag, id1, false)
	if err != nil {
		t.Errorf("getCallback error for tag %q and ID %d: %+v", tag, id1, err)
	}

	if reflect.ValueOf(cb).Pointer() != reflect.ValueOf(received).Pointer() {
		t.Errorf("Wrong callback.\nexpected: %p\nreceived: %p", cb, received)
	}

	// Add new callback under the same tag but with deleteCB set to true and
	// check that it is returned by getCallback and that it was deleted from the
	// map while id1 was not
	id2 := uint64(56)
	cb = func([]byte) {}
	m.callbacks[tag][id2] = cb

	received, err = m.getCallback(tag, id2, true)
	if err != nil {
		t.Errorf("getCallback error for tag %q and ID %d: %+v", tag, id2, err)
	}

	if reflect.ValueOf(cb).Pointer() != reflect.ValueOf(received).Pointer() {
		t.Errorf("Wrong callback.\nexpected: %p\nreceived: %p", cb, received)
	}

	received, err = m.getCallback(tag, id1, false)
	if err != nil {
		t.Errorf("getCallback error for tag %q and ID %d: %+v", tag, id1, err)
	}

	received, err = m.getCallback(tag, id2, true)
	if err == nil {
		t.Errorf("getCallback did not get error when trying to get deleted "+
			"callback for tag %q and ID %d", tag, id2)
	}
}

// Tests that Manager.RegisterCallback registers a callback that is then called
// by Manager.receiveMessage.
func TestManager_RegisterCallback(t *testing.T) {
	m := &Manager{callbacks: make(map[Tag]map[uint64]ReceptionCallback)}

	msg := Message{Tag: readyTag, ID: initID}
	cbChan := make(chan struct{})
	cb := func([]byte) { cbChan <- struct{}{} }
	m.RegisterCallback(msg.Tag, cb)

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to JSON marshal Message: %+v", err)
	}

	go func() {
		select {
		case <-cbChan:
		case <-time.After(10 * time.Millisecond):
			t.Error("Timed out waiting for callback to be called.")
		}
	}()

	err = m.receiveMessage(data)
	if err != nil {
		t.Errorf("Failed to receive message: %+v", err)
	}
}

// Tests that Manager.registerReplyCallback registers a callback that is then
// called by Manager.receiveMessage.
func TestManager_registerReplyCallback(t *testing.T) {
	m := &Manager{
		callbacks:   make(map[Tag]map[uint64]ReceptionCallback),
		responseIDs: make(map[Tag]uint64),
	}

	msg := Message{Tag: readyTag, ID: 5}
	cbChan := make(chan struct{})
	cb := func([]byte) { cbChan <- struct{}{} }
	m.registerReplyCallback(msg.Tag, cb)
	m.callbacks[msg.Tag] = map[uint64]ReceptionCallback{msg.ID: cb}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to JSON marshal Message: %+v", err)
	}

	go func() {
		select {
		case <-cbChan:
		case <-time.After(10 * time.Millisecond):
			t.Error("Timed out waiting for callback to be called.")
		}
	}()

	err = m.receiveMessage(data)
	if err != nil {
		t.Errorf("Failed to receive message: %+v", err)
	}
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

////////////////////////////////////////////////////////////////////////////////
// Javascript Call Wrappers                                                   //
////////////////////////////////////////////////////////////////////////////////

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
