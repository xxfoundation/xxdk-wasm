////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package worker

/*
// Unit test of initManager.
func Test_initManager(t *testing.T) {
	expected := &Manager{
		senderCallbacks:   make(map[Tag]map[uint64]SenderCallback),
		receiverCallbacks: make(map[Tag]ReceiverCallback),
		responseIDs:       make(map[Tag]uint64),
		receiveQueue:      make(chan js.Value, receiveQueueChanSize),
		quit:              make(chan struct{}),
		name:              "name",
		messageLogging:    true,
	}

	received := initManager(expected.name, expected.messageLogging)

	received.receiveQueue = expected.receiveQueue
	received.quit = expected.quit
	if !reflect.DeepEqual(expected, received) {
		t.Errorf("Unexpected Manager.\nexpected: %+v\nreceived: %+v",
			expected, received)
	}
}

// Tests Manager.processReceivedMessage calls the expected callback.
func TestManager_processReceivedMessage(t *testing.T) {
	m := initManager("", true)

	msg := Message{Tag: readyTag, ID: 5}
	cbChan := make(chan struct{})
	cb := func([]byte, func([]byte)) { cbChan <- struct{}{} }
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

	err = m.processReceivedMessage(data)
	if err != nil {
		t.Errorf("Failed to receive message: %+v", err)
	}

	msg = Message{Tag: "tag", Response: true}
	cbChan = make(chan struct{})
	cb2 := func([]byte) { cbChan <- struct{}{} }
	m.registerSenderCallback(msg.Tag, cb2)

	go func() {
		select {
		case <-cbChan:
		case <-time.After(10 * time.Millisecond):
			t.Error("Timed out waiting for callback to be called.")
		}
	}()

	err = m.processReceivedMessage(data)
	if err != nil {
		t.Errorf("Failed to receive message: %+v", err)
	}
}

// Tests Manager.getSenderCallback returns the expected callback and deletes it.
func TestManager_getSenderCallback(t *testing.T) {
	m := initManager("", true)

	expected := make(map[Tag]map[uint64]SenderCallback)
	for i := 0; i < 5; i++ {
		tag := Tag("tag" + strconv.Itoa(i))
		expected[tag] = make(map[uint64]SenderCallback)
		for j := 0; j < 10; j++ {
			cb := func([]byte) {}
			id := m.registerSenderCallback(tag, cb)
			expected[tag][id] = cb
		}
	}

	for tag, callbacks := range expected {
		for id, callback := range callbacks {
			received, err := m.getSenderCallback(tag, id)
			if err != nil {
				t.Errorf("Error getting callback for tag %q and ID %d: %+v",
					tag, id, err)
			}

			if reflect.ValueOf(callback).Pointer() != reflect.ValueOf(received).Pointer() {
				t.Errorf("Wrong callback for tag %q and ID %d."+
					"\nexpected: %p\nreceived: %p", tag, id, callback, received)
			}

			// Check that the callback was deleted
			if received, err = m.getSenderCallback(tag, id); err == nil {
				t.Errorf("Did not get error when for callback that should be "+
					"deleted for tag %q and ID %d: %p", tag, id, received)
			}
		}
		if callbacks, exists := m.senderCallbacks[tag]; exists {
			t.Errorf("Empty map for tag %s not deleted: %+v", tag, callbacks)
		}
	}
}

// Tests Manager.getReceiverCallback returns the expected callback.
func TestManager_getReceiverCallback(t *testing.T) {
	m := initManager("", true)

	expected := make(map[Tag]ReceiverCallback)
	for i := 0; i < 5; i++ {
		tag := Tag("tag" + strconv.Itoa(i))
		cb := func([]byte, func([]byte)) {}
		m.RegisterCallback(tag, cb)
		expected[tag] = cb
	}

	for tag, callback := range expected {
		received, err := m.getReceiverCallback(tag)
		if err != nil {
			t.Errorf("Error getting callback for tag %q: %+v", tag, err)
		}

		if reflect.ValueOf(callback).Pointer() != reflect.ValueOf(received).Pointer() {
			t.Errorf("Wrong callback for tag %q."+
				"\nexpected: %p\nreceived: %p", tag, callback, received)
		}
	}
}

// Tests that Manager.RegisterCallback registers a callback that is then called
// by Manager.processReceivedMessage.
func TestManager_RegisterCallback(t *testing.T) {
	m := initManager("", true)

	msg := Message{Tag: readyTag, ID: initID}
	cbChan := make(chan struct{})
	cb := func([]byte, func([]byte)) { cbChan <- struct{}{} }
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

	err = m.processReceivedMessage(data)
	if err != nil {
		t.Errorf("Failed to receive message: %+v", err)
	}
}

// Tests that Manager.registerSenderCallback registers a callback that is then
// called by Manager.processReceivedMessage.
func TestManager_registerSenderCallback(t *testing.T) {
	m := initManager("", true)

	msg := Message{Tag: readyTag, Response: true}
	cbChan := make(chan struct{})
	cb := func([]byte) { cbChan <- struct{}{} }
	msg.ID = m.registerSenderCallback(msg.Tag, cb)

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

	err = m.processReceivedMessage(data)
	if err != nil {
		t.Errorf("Failed to receive message: %+v", err)
	}
}

// Tests that Manager.getNextID returns the expected ID for various Tags.
func TestManager_getNextID(t *testing.T) {
	m := initManager("", true)

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

				v := opts["type"].(string)
				if v != workerType {
					t.Errorf("Unexpected type (%d, %d, %d)."+
						"\nexpected: %s\nreceived: %s", i, j, k, workerType, v)
				}

				v = opts["credentials"].(string)
				if v != credentials {
					t.Errorf("Unexpected credentials (%d, %d, %d)."+
						"\nexpected: %s\nreceived: %s", i, j, k, credentials, v)
				}

				v = opts["name"].(string)
				if v != name {
					t.Errorf("Unexpected name (%d, %d, %d)."+
						"\nexpected: %s\nreceived: %s", i, j, k, name, v)
				}
			}
		}
	}
}
*/
