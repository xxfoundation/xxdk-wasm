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
	"testing"
	"time"
)

// Tests that ThreadManager.processReceivedMessage calls the expected callback.
func TestThreadManager_processReceivedMessage(t *testing.T) {
	tm := &ThreadManager{callbacks: make(map[Tag]ThreadReceptionCallback)}

	msg := Message{Tag: readyTag, ID: 5}
	cbChan := make(chan struct{}, 1)
	cb := func([]byte) ([]byte, error) { cbChan <- struct{}{}; return nil, nil }
	tm.callbacks[msg.Tag] = cb

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to JSON marshal Message: %+v", err)
	}

	go func() {
		err = tm.processReceivedMessage(data)
		if err != nil {
			t.Errorf("Failed to receive message: %+v", err)
		}
	}()

	select {
	case <-cbChan:
	case <-time.After(10 * time.Millisecond):
		t.Error("Timed out waiting for callback to be called.")
	}
}

// Tests that ThreadManager.RegisterCallback registers a callback that is then
// called by ThreadManager.processReceivedMessage.
func TestThreadManager_RegisterCallback(t *testing.T) {
	tm := &ThreadManager{callbacks: make(map[Tag]ThreadReceptionCallback)}

	msg := Message{Tag: readyTag, ID: 5}
	cbChan := make(chan struct{}, 1)
	cb := func([]byte) ([]byte, error) { cbChan <- struct{}{}; return nil, nil }
	tm.RegisterCallback(msg.Tag, cb)

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to JSON marshal Message: %+v", err)
	}

	go func() {
		err = tm.processReceivedMessage(data)
		if err != nil {
			t.Errorf("Failed to receive message: %+v", err)
		}
	}()

	select {
	case <-cbChan:
	case <-time.After(10 * time.Millisecond):
		t.Error("Timed out waiting for callback to be called.")
	}
}
