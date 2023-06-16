////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package main

import (
	"bytes"
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"gitlab.com/elixxir/client/v4/cmix/rounds"
	"gitlab.com/elixxir/client/v4/dm"
	"gitlab.com/elixxir/crypto/message"
	"gitlab.com/elixxir/wasm-utils/utils"
	"gitlab.com/elixxir/xxdk-wasm/indexedDb/impl"
	"gitlab.com/xx_network/primitives/id"
	"os"
	"syscall/js"
	"testing"
	"time"

	jww "github.com/spf13/jwalterweatherman"
)

func dummyReceivedMessageCB(uint64, ed25519.PublicKey, bool, bool) {}

func TestMain(m *testing.M) {
	jww.SetStdoutThreshold(jww.LevelDebug)
	os.Exit(m.Run())
}

// Test simple receive of a new message for a new conversation.
func TestImpl_Receive(t *testing.T) {
	m, err := newWASMModel("TestImpl_Receive", nil,
		dummyReceivedMessageCB)
	if err != nil {
		t.Fatal(err.Error())
	}

	testString := "test"
	testBytes := []byte(testString)
	partnerPubKey := ed25519.PublicKey(testBytes)
	testRound := id.Round(10)

	// Can use ChannelMessageID for ease, doesn't matter here
	testMsgId := message.DeriveChannelMessageID(&id.ID{1}, uint64(testRound), testBytes)

	// Receive a test message
	uuid := m.Receive(testMsgId, testString, testBytes,
		partnerPubKey, partnerPubKey, 0, 0, time.Now(),
		rounds.Round{ID: testRound}, dm.TextType, dm.Received)
	if uuid == 0 {
		t.Fatalf("Expected non-zero message uuid")
	}
	jww.DEBUG.Printf("Received test message: %d", uuid)

	// First, we expect a conversation to be created
	testConvo := m.GetConversation(partnerPubKey)
	if testConvo == nil {
		t.Fatalf("Expected conversation to be created")
	}
	// Spot check a conversation attribute
	if testConvo.Nickname != testString {
		t.Fatalf("Expected conversation nickname %s, got %s",
			testString, testConvo.Nickname)
	}

	// Next, we expect the message to be created
	testMessageObj, err := impl.Get(m.db, messageStoreName, js.ValueOf(uuid))
	if err != nil {
		t.Fatalf(err.Error())
	}
	testMessage := &Message{}
	err = json.Unmarshal([]byte(utils.JsToJson(testMessageObj)), testMessage)
	if err != nil {
		t.Fatalf(err.Error())
	}
	// Spot check a message attribute
	if !bytes.Equal(testMessage.SenderPubKey, partnerPubKey) {
		t.Fatalf("Expected message attibutes to match, expected %v got %v",
			partnerPubKey, testMessage.SenderPubKey)
	}
}

// Test happy path. Insert some conversations and check they exist.
func TestImpl_GetConversations(t *testing.T) {
	m, err := newWASMModel("TestImpl_GetConversations", nil,
		dummyReceivedMessageCB)
	if err != nil {
		t.Fatal(err.Error())
	}
	numTestConvo := 10

	// Insert a test convo
	for i := 0; i < numTestConvo; i++ {
		testBytes := []byte(fmt.Sprintf("%d", i))
		testPubKey := ed25519.PublicKey(testBytes)
		err = m.upsertConversation("test", testPubKey,
			uint32(i), uint8(i), nil)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	results := m.GetConversations()
	if len(results) != numTestConvo {
		t.Fatalf("Expected %d convos, got %d", numTestConvo, len(results))
	}

	for i, convo := range results {
		if convo.Token != uint32(i) {
			t.Fatalf("Expected %d convo token, got %d", i, convo.Token)
		}
		if convo.CodesetVersion != uint8(i) {
			t.Fatalf("Expected %d convo codeset, got %d",
				i, convo.CodesetVersion)
		}
	}
}

// Test happy path toggling between blocked/unblocked in a Conversation.
func TestWasmModel_BlockSender(t *testing.T) {
	m, err := newWASMModel("TestWasmModel_BlockSender", nil, dummyReceivedMessageCB)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Insert a test convo
	testPubKey := ed25519.PublicKey{}
	err = m.upsertConversation("test", testPubKey, 0, 0, nil)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Default to unblocked
	result := m.GetConversation(testPubKey)
	if result.BlockedTimestamp != nil {
		t.Fatal("Expected blocked to be false")
	}

	// Now toggle blocked
	m.BlockSender(testPubKey)
	result = m.GetConversation(testPubKey)
	if result.BlockedTimestamp == nil {
		t.Fatal("Expected blocked to be true")
	}

	// Now toggle blocked again
	m.UnblockSender(testPubKey)
	result = m.GetConversation(testPubKey)
	if result.BlockedTimestamp != nil {
		t.Fatal("Expected blocked to be false")
	}
}

// Test failed and successful deletes
func TestWasmModel_DeleteMessage(t *testing.T) {
	m, err := newWASMModel("TestWasmModel_DeleteMessage", nil, dummyReceivedMessageCB)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Insert test message
	testBytes := []byte("test")
	testBadBytes := []byte("uwu")
	testMsgId := message.DeriveChannelMessageID(&id.ID{1}, 0, testBytes)
	testMsg := &Message{
		MessageID:          testMsgId.Marshal(),
		ConversationPubKey: testBytes,
		ParentMessageID:    nil,
		Timestamp:          time.Now(),
		SenderPubKey:       testBytes,
		CodesetVersion:     5,
		Status:             5,
		Text:               "",
		Type:               5,
		Round:              5,
	}
	_, err = m.upsertMessage(testMsg)
	require.NoError(t, err)

	// Non-matching pub key, should fail to delete
	require.False(t, m.DeleteMessage(testMsgId, testBadBytes))

	// Correct pub key, should have deleted
	require.True(t, m.DeleteMessage(testMsgId, testBytes))
}
