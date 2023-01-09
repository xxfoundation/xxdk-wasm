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

func TestNewManager(t *testing.T) {
}

func TestManager_SendMessage(t *testing.T) {
}

func TestManager_receiveMessage(t *testing.T) {
}

func TestManager_getHandler(t *testing.T) {
}

func TestManager_RegisterHandler(t *testing.T) {
}

// Tests that Manager.getNextID returns the expected ID for various Tags.
func TestManager_getNextID(t *testing.T) {
	m := &Manager{
		callbacks:   make(map[Tag]map[uint64]ReceptionCallback),
		responseIDs: make(map[Tag]uint64),
	}

	for _, tag := range []Tag{
		ReadyTag, NewWASMEventModelTag, EncryptionStatusTag,
		StoreDatabaseNameTag, JoinChannelTag, LeaveChannelTag,
		ReceiveMessageTag, ReceiveReplyTag, ReceiveReactionTag,
		UpdateFromUUIDTag, UpdateFromMessageIDTag, GetMessageTag,
		DeleteMessageTag, ReceiveTag, ReceiveTextTag, UpdateSentStatusTag,
	} {
		id := m.getNextID(tag)
		if id != InitID {
			t.Errorf("ID for new tag %q is not InitID."+
				"\nexpected: %d\nreceived: %d", tag, InitID, id)
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

func TestManager_addEventListeners(t *testing.T) {
}

func TestManager_postMessage(t *testing.T) {
}

func Test_newWorkerOptions(t *testing.T) {
}
