////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package indexedDb

import (
	"encoding/json"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/client/channels"
	cryptoBroadcast "gitlab.com/elixxir/crypto/broadcast"
	"gitlab.com/elixxir/crypto/channel"
	"gitlab.com/xx_network/primitives/id"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	jww.SetStdoutThreshold(jww.LevelDebug)
	os.Exit(m.Run())
}

// Test wasmModel.UpdateSentStatus happy path and ensure fields don't change.
func TestWasmModel_UpdateSentStatus(t *testing.T) {
	testString := "test"
	testMsgId := channel.MakeMessageID([]byte(testString))
	eventModel, err := newWasmModel(testString)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	// Store a test message
	testMsg := buildMessage([]byte(testString), testMsgId.Bytes(),
		nil, testString, testString, nil, time.Now(),
		time.Second, channels.Sent)
	err = eventModel.receiveHelper(testMsg)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	// Ensure one message is stored
	results, err := eventModel.dump(messageStoreName)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	if len(results) != 1 {
		t.Fatalf("Expected 1 message to exist")
	}

	// Update the sentStatus
	expectedStatus := channels.Failed
	eventModel.UpdateSentStatus(testMsgId, expectedStatus)

	// Check the resulting status
	results, err = eventModel.dump(messageStoreName)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	if len(results) != 1 {
		t.Fatalf("Expected 1 message to exist")
	}
	resultMsg := &Message{}
	err = json.Unmarshal([]byte(results[0]), resultMsg)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	if resultMsg.Status != uint8(expectedStatus) {
		t.Fatalf("Unexpected Status: %v", resultMsg.Status)
	}

	// Make sure other fields didn't change
	if resultMsg.SenderUsername != testString {
		t.Fatalf("Unexpected SenderUsername: %v", resultMsg.SenderUsername)
	}
}

// Smoke test wasmModel.JoinChannel/wasmModel.LeaveChannel happy paths.
func TestWasmModel_JoinChannel_LeaveChannel(t *testing.T) {
	eventModel, err := newWasmModel("test")
	if err != nil {
		t.Fatalf("%+v", err)
	}

	testChannel := &cryptoBroadcast.Channel{
		ReceptionID: id.NewIdFromString("test", id.Generic, t),
		Name:        "test",
		Description: "test",
		Salt:        nil,
	}
	testChannel2 := &cryptoBroadcast.Channel{
		ReceptionID: id.NewIdFromString("test2", id.Generic, t),
		Name:        "test2",
		Description: "test2",
		Salt:        nil,
	}
	eventModel.JoinChannel(testChannel)
	eventModel.JoinChannel(testChannel2)
	results, err := eventModel.dump(channelsStoreName)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	if len(results) != 2 {
		t.Fatalf("Expected 2 channels to exist")
	}
	eventModel.LeaveChannel(testChannel.ReceptionID)
	results, err = eventModel.dump(channelsStoreName)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	if len(results) != 1 {
		t.Fatalf("Expected 1 channels to exist")
	}
}
