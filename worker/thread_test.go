////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package worker

/*
// Unit test of initThreadManager.
func Test_initThreadManager(t *testing.T) {
	expected := &ThreadManager{
		senderCallbacks:   make(map[Tag]map[uint64]SenderCallback),
		receiverCallbacks: make(map[Tag]ReceiverCallback),
		responseIDs:       make(map[Tag]uint64),
		receiveQueue:      make(chan js.Value, receiveQueueChanSize),
		quit:              make(chan struct{}),
		name:              "name",
		messageLogging:    true,
	}

	received := initThreadManager(expected.name, expected.messageLogging)

	received.receiveQueue = expected.receiveQueue
	received.quit = expected.quit
	if !reflect.DeepEqual(expected, received) {
		t.Errorf("Unexpected ThreadManager.\nexpected: %+v\nreceived: %+v",
			expected, received)
	}
}

// Tests that ThreadManager.processReceivedMessage calls the expected callback.
func TestThreadManager_processReceivedMessage(t *testing.T) {
	tm := initThreadManager("", true)

	msg := Message{Tag: readyTag, ID: 5}
	cbChan := make(chan struct{})
	cb := func([]byte, func([]byte)) { cbChan <- struct{}{} }
	tm.receiverCallbacks[msg.Tag] = cb

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

	err = tm.processReceivedMessage(data)
	if err != nil {
		t.Errorf("Failed to receive message: %+v", err)
	}
}

// Tests ThreadManager.getSenderCallback returns the expected callback and
// deletes it.
func TestThreadManager_getSenderCallback(t *testing.T) {
	tm := initThreadManager("", true)

	expected := make(map[Tag]map[uint64]SenderCallback)
	for i := 0; i < 5; i++ {
		tag := Tag("tag" + strconv.Itoa(i))
		expected[tag] = make(map[uint64]SenderCallback)
		for j := 0; j < 10; j++ {
			cb := func([]byte) {}
			id := tm.registerSenderCallback(tag, cb)
			expected[tag][id] = cb
		}
	}

	for tag, callbacks := range expected {
		for id, callback := range callbacks {
			received, err := tm.getSenderCallback(tag, id)
			if err != nil {
				t.Errorf("Error getting callback for tag %q and ID %d: %+v",
					tag, id, err)
			}

			if reflect.ValueOf(callback).Pointer() != reflect.ValueOf(received).Pointer() {
				t.Errorf("Wrong callback for tag %q and ID %d."+
					"\nexpected: %p\nreceived: %p", tag, id, callback, received)
			}

			// Check that the callback was deleted
			if received, err = tm.getSenderCallback(tag, id); err == nil {
				t.Errorf("Did not get error when for callback that should be "+
					"deleted for tag %q and ID %d: %p", tag, id, received)
			}
		}
		if callbacks, exists := tm.senderCallbacks[tag]; exists {
			t.Errorf("Empty map for tag %s not deleted: %+v", tag, callbacks)
		}
	}
}

// Tests ThreadManager.getReceiverCallback returns the expected callback.
func TestThreadManager_getReceiverCallback(t *testing.T) {
	tm := initThreadManager("", true)

	expected := make(map[Tag]ReceiverCallback)
	for i := 0; i < 5; i++ {
		tag := Tag("tag" + strconv.Itoa(i))
		cb := func([]byte, func([]byte)) {}
		tm.RegisterCallback(tag, cb)
		expected[tag] = cb
	}

	for tag, callback := range expected {
		received, err := tm.getReceiverCallback(tag)
		if err != nil {
			t.Errorf("Error getting callback for tag %q: %+v", tag, err)
		}

		if reflect.ValueOf(callback).Pointer() != reflect.ValueOf(received).Pointer() {
			t.Errorf("Wrong callback for tag %q."+
				"\nexpected: %p\nreceived: %p", tag, callback, received)
		}
	}
}

// Tests that ThreadManager.RegisterCallback registers a callback that is then
// called by ThreadManager.processReceivedMessage.
func TestThreadManager_RegisterCallback(t *testing.T) {
	tm := initThreadManager("", true)

	msg := Message{Tag: readyTag, ID: 5}
	cbChan := make(chan struct{})
	cb := func([]byte, func([]byte)) { cbChan <- struct{}{} }
	tm.RegisterCallback(msg.Tag, cb)

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

	err = tm.processReceivedMessage(data)
	if err != nil {
		t.Errorf("Failed to receive message: %+v", err)
	}
}

// Tests that ThreadManager.registerSenderCallback registers a callback that is
// then called by ThreadManager.processReceivedMessage.
func TestThreadManager_registerSenderCallback(t *testing.T) {
	tm := initThreadManager("", true)

	msg := Message{Tag: readyTag, Response: true}
	cbChan := make(chan struct{})
	cb := func([]byte) { cbChan <- struct{}{} }
	msg.ID = tm.registerSenderCallback(msg.Tag, cb)

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

	err = tm.processReceivedMessage(data)
	if err != nil {
		t.Errorf("Failed to receive message: %+v", err)
	}
}

// Tests that ThreadManager.getNextID returns the expected ID for various Tags.
func TestThreadManager_getNextID(t *testing.T) {
	tm := initThreadManager("", true)

	for _, tag := range []Tag{readyTag, "test", "A", "B", "C"} {
		id := tm.getNextID(tag)
		if id != initID {
			t.Errorf("ID for new tag %q is not initID."+
				"\nexpected: %d\nreceived: %d", tag, initID, id)
		}

		for j := uint64(1); j < 100; j++ {
			id = tm.getNextID(tag)
			if id != j {
				t.Errorf("Unexpected ID for tag %q."+
					"\nexpected: %d\nreceived: %d", tag, j, id)
			}
		}
	}
}
*/
