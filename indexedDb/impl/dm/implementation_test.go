////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package main

import (
	"crypto/ed25519"
	jww "github.com/spf13/jwalterweatherman"
	"os"
	"testing"
)

func dummyReceivedMessageCB(uint64, ed25519.PublicKey, bool) {}
func dummyStoreDatabaseName(string) error                    { return nil }
func dummyStoreEncryptionStatus(_ string, encryptionStatus bool) (bool, error) {
	return encryptionStatus, nil
}

func TestMain(m *testing.M) {
	jww.SetStdoutThreshold(jww.LevelDebug)
	os.Exit(m.Run())
}

// Test happy path toggling between blocked/unblocked in a Conversation.
func TestWasmModel_BlockSender(t *testing.T) {
	m, err := newWASMModel("test", nil,
		dummyReceivedMessageCB, dummyStoreDatabaseName, dummyStoreEncryptionStatus)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Insert a test convo
	testPubKey := ed25519.PublicKey{}
	err = m.joinConversation("test", testPubKey, 0, 0)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Default to unblocked
	result := m.GetConversation(testPubKey)
	if result.Blocked {
		t.Fatal("Expected blocked to be false")
	}

	// Now toggle blocked
	m.BlockSender(testPubKey)
	result = m.GetConversation(testPubKey)
	if !result.Blocked {
		t.Fatal("Expected blocked to be true")
	}

	// Now toggle blocked again
	m.UnblockSender(testPubKey)
	result = m.GetConversation(testPubKey)
	if result.Blocked {
		t.Fatal("Expected blocked to be false")
	}
}
