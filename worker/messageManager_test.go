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
	"strconv"
	"syscall/js"
	"testing"
	"time"

	"github.com/hack-pad/safejs"
	"gitlab.com/elixxir/wasm-utils/utils"
)

func TestNewMessageManager(t *testing.T) {
}

// Unit test of initMessageManager.
func Test_initMessageManager(t *testing.T) {
	expected := &MessageManager{
		senderCallbacks:   make(map[Tag]map[uint64]SenderCallback),
		receiverCallbacks: make(map[Tag]ReceiverCallback),
		responseIDs:       make(map[Tag]uint64),
		messageChannelCB:  make(map[string]NewPortCallback),
		quit:              make(chan struct{}),
		name:              "name",
		Params:            DefaultParams(),
	}

	received := initMessageManager(expected.name, expected.Params)

	received.quit = expected.quit
	if !reflect.DeepEqual(expected, received) {
		t.Errorf("Unexpected MessageManager.\nexpected: %+v\nreceived: %+v",
			expected, received)
	}
}

func TestMessageManager_Send(t *testing.T) {
}

func TestMessageManager_SendTimeout(t *testing.T) {
}

func TestMessageManager_SendNoResponse(t *testing.T) {
}

func TestMessageManager_sendMessage(t *testing.T) {
}

func TestMessageManager_sendResponse(t *testing.T) {
}

func TestMessageManager_messageReception(t *testing.T) {
}

// Tests MessageManager.processReceivedMessage calls the expected callback.
func TestMessageManager_processReceivedMessage(t *testing.T) {
	mm := initMessageManager("", DefaultParams())

	msg := Message{Tag: readyTag, ID: 5}
	cbChan := make(chan struct{})
	cb := func([]byte, func([]byte)) { cbChan <- struct{}{} }
	mm.RegisterCallback(msg.Tag, cb)

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

	err = mm.processReceivedMessage(data)
	if err != nil {
		t.Errorf("Failed to receive message: %+v", err)
	}

	msg = Message{Tag: "tag", Response: true}
	cbChan = make(chan struct{})
	cb2 := func([]byte) { cbChan <- struct{}{} }
	mm.registerSenderCallback(msg.Tag, cb2)

	go func() {
		select {
		case <-cbChan:
		case <-time.After(10 * time.Millisecond):
			t.Error("Timed out waiting for callback to be called.")
		}
	}()

	err = mm.processReceivedMessage(data)
	if err != nil {
		t.Errorf("Failed to receive message: %+v", err)
	}

	time.Sleep(15 * time.Millisecond)
}

// Tests MessageManager.processReceivedPort calls the expected callback.
func TestMessageManager_processReceivedPort(t *testing.T) {
	mm := initMessageManager("", DefaultParams())

	cbChan := make(chan string)
	cb := func(port js.Value, channelName string) { cbChan <- channelName }
	key := "testKey"
	channelName := "channel"
	mm.RegisterMessageChannelCallback(key, cb)

	mc, err := NewMessageChannel()
	if err != nil {
		t.Fatal(err)
	}
	port1, err := mc.Port1()
	if err != nil {
		t.Fatalf("Failed to get port1: %+v", err)
	}

	obj := map[string]any{
		"port":    safejs.Unsafe(port1.Value),
		"channel": utils.CopyBytesToJS([]byte(channelName)),
		"key":     utils.CopyBytesToJS([]byte(key))}

	go func() {
		select {
		case name := <-cbChan:
			if channelName != name {
				t.Errorf("Received incorrect channel name."+
					"\nexpected: %q\nrecieved: %q", channelName, name)
			}
		case <-time.After(10 * time.Millisecond):
			t.Error("Timed out waiting for callback to be called.")
		}
	}()

	data := js.ValueOf(obj)
	err = mm.processReceivedPort(data.Get("port"), data)
	if err != nil {
		t.Errorf("Failed to receive message: %+v", err)
	}
	time.Sleep(15 * time.Millisecond)
}

// Tests that MessageManager.RegisterCallback registers a callback that is then
// called by MessageManager.processReceivedMessage.
func TestMessageManager_RegisterCallback(t *testing.T) {
	mm := initMessageManager("", DefaultParams())

	msg := Message{Tag: readyTag, ID: initID}
	cbChan := make(chan struct{})
	cb := func([]byte, func([]byte)) { cbChan <- struct{}{} }
	mm.RegisterCallback(msg.Tag, cb)

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

	err = mm.processReceivedMessage(data)
	if err != nil {
		t.Errorf("Failed to receive message: %+v", err)
	}
	time.Sleep(15 * time.Millisecond)
}

// Tests MessageManager.getReceiverCallback returns the expected callback.
func TestMessageManager_getReceiverCallback(t *testing.T) {
	mm := initMessageManager("", DefaultParams())

	expected := make(map[Tag]ReceiverCallback)
	for i := 0; i < 5; i++ {
		tag := Tag("tag" + strconv.Itoa(i))
		cb := func([]byte, func([]byte)) {}
		mm.RegisterCallback(tag, cb)
		expected[tag] = cb
	}

	for tag, cb := range expected {
		r, err := mm.getReceiverCallback(tag)
		if err != nil {
			t.Errorf("Error getting callback for tag %q: %+v", tag, err)
		}

		if reflect.ValueOf(cb).Pointer() != reflect.ValueOf(r).Pointer() {
			t.Errorf("Wrong callback for tag %q."+
				"\nexpected: %p\nreceived: %p", tag, cb, r)
		}
	}
}

// Tests that MessageManager.registerSenderCallback registers a callback that is
// then called by MessageManager.processReceivedMessage.
func TestMessageManager_registerSenderCallback(t *testing.T) {
	mm := initMessageManager("", DefaultParams())

	msg := Message{Tag: readyTag, Response: true}
	cbChan := make(chan struct{})
	cb := func([]byte) { cbChan <- struct{}{} }
	msg.ID = mm.registerSenderCallback(msg.Tag, cb)

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

	err = mm.processReceivedMessage(data)
	if err != nil {
		t.Errorf("Failed to receive message: %+v", err)
	}
	time.Sleep(15 * time.Millisecond)
}

// Tests MessageManager.getSenderCallback returns the expected callback and
// deletes it.
func TestMessageManager_getSenderCallback(t *testing.T) {
	mm := initMessageManager("", DefaultParams())

	expected := make(map[Tag]map[uint64]SenderCallback)
	for i := 0; i < 5; i++ {
		tag := Tag("tag" + strconv.Itoa(i))
		expected[tag] = make(map[uint64]SenderCallback)
		for j := 0; j < 10; j++ {
			cb := func([]byte) {}
			id := mm.registerSenderCallback(tag, cb)
			expected[tag][id] = cb
		}
	}

	for tag, callbacks := range expected {
		for id, cb := range callbacks {
			r, err := mm.getSenderCallback(tag, id)
			if err != nil {
				t.Errorf("Error getting callback for tag %q and ID %d: %+v",
					tag, id, err)
			}

			if reflect.ValueOf(cb).Pointer() != reflect.ValueOf(r).Pointer() {
				t.Errorf("Wrong callback for tag %q and ID %d."+
					"\nexpected: %p\nreceived: %p", tag, id, cb, r)
			}

			// Check that the callback was deleted
			if r, err = mm.getSenderCallback(tag, id); err == nil {
				t.Errorf("Did not get error when for callback that should be "+
					"deleted for tag %q and ID %d: %p", tag, id, r)
			}
		}
		if callbacks, exists := mm.senderCallbacks[tag]; exists {
			t.Errorf("Empty map for tag %s not deleted: %+v", tag, callbacks)
		}
	}
}

// Tests that MessageManager.RegisterMessageChannelCallback registers a callback
// that is then called by MessageManager.processReceivedPort.
func TestMessageManager_RegisterMessageChannelCallback(t *testing.T) {
	mm := initMessageManager("", DefaultParams())

	cbChan := make(chan string)
	cb := func(port js.Value, channelName string) { cbChan <- channelName }
	key := "testKey"
	channelName := "channel"
	mm.RegisterMessageChannelCallback(key, cb)

	mc, err := NewMessageChannel()
	if err != nil {
		t.Fatal(err)
	}
	port1, err := mc.Port1()
	if err != nil {
		t.Fatalf("Failed to get port1: %+v", err)
	}

	obj := map[string]any{
		"port":    safejs.Unsafe(port1.Value),
		"channel": utils.CopyBytesToJS([]byte(channelName)),
		"key":     utils.CopyBytesToJS([]byte(key))}

	go func() {
		select {
		case name := <-cbChan:
			if channelName != name {
				t.Errorf("Received incorrect channel name."+
					"\nexpected: %q\nrecieved: %q", channelName, name)
			}
		case <-time.After(10 * time.Millisecond):
			t.Error("Timed out waiting for callback to be called.")
		}
	}()

	data := js.ValueOf(obj)
	err = mm.processReceivedPort(data.Get("port"), data)
	if err != nil {
		t.Errorf("Failed to receive message: %+v", err)
	}
	time.Sleep(15 * time.Millisecond)
}

func TestMessageManager_Stop(t *testing.T) {
}

// Tests that MessageManager.getNextID returns the expected ID for various Tags.
func TestMessageManager_getNextID(t *testing.T) {
	mm := initMessageManager("", DefaultParams())

	for _, tag := range []Tag{readyTag, "test", "A", "B", "C"} {
		id := mm.getNextID(tag)
		if id != initID {
			t.Errorf("ID for new tag %q is not initID."+
				"\nexpected: %d\nreceived: %d", tag, initID, id)
		}

		for j := initID + 1; j < 100; j++ {
			id = mm.getNextID(tag)
			if id != j {
				t.Errorf("Unexpected ID for tag %q."+
					"\nexpected: %d\nreceived: %d", tag, j, id)
			}
		}
	}
}
