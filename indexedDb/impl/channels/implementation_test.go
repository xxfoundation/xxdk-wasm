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
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/hack-pad/go-indexeddb/idb"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/stretchr/testify/require"

	"gitlab.com/elixxir/client/v4/channels"
	cft "gitlab.com/elixxir/client/v4/channelsFileTransfer"
	"gitlab.com/elixxir/client/v4/cmix/rounds"
	cryptoBroadcast "gitlab.com/elixxir/crypto/broadcast"
	"gitlab.com/elixxir/crypto/fileTransfer"
	idbCrypto "gitlab.com/elixxir/crypto/indexedDb"
	"gitlab.com/elixxir/crypto/message"
	"gitlab.com/elixxir/wasm-utils/storage"
	"gitlab.com/elixxir/xxdk-wasm/indexedDb/impl"
	"gitlab.com/xx_network/crypto/csprng"
	"gitlab.com/xx_network/primitives/id"
	"gitlab.com/xx_network/primitives/netTime"
)

func TestMain(m *testing.M) {
	jww.SetStdoutThreshold(jww.LevelDebug)
	os.Exit(m.Run())
}

type dummyCbs struct{}

func (c *dummyCbs) EventUpdate(int64, []byte) {}

// Happy path test for receiving, updating, getting, and deleting a File.
func TestWasmModel_ReceiveFile(t *testing.T) {
	testString := "TestWasmModel_ReceiveFile"
	m, err := newWASMModel(testString, nil, &dummyCbs{})
	if err != nil {
		t.Fatal(err)
	}

	testTs := time.Now()
	testBytes := []byte(testString)
	testStatus := cft.Downloading

	// Insert a test row
	fId := fileTransfer.NewID(testBytes)
	err = m.ReceiveFile(fId, testBytes, testBytes, testTs, testStatus)
	if err != nil {
		t.Fatal(err)
	}

	// Attempt to get stored row
	storedFile, err := m.GetFile(fId)
	if err != nil {
		t.Fatal(err)
	}
	// Spot check stored attribute
	if !bytes.Equal(storedFile.Link, testBytes) {
		t.Fatalf("Got unequal FileLink values")
	}

	// Attempt to updated stored row
	newTs := time.Now()
	newBytes := []byte("test")
	newStatus := cft.Complete
	err = m.UpdateFile(fId, nil, newBytes, &newTs, &newStatus)
	if err != nil {
		t.Fatal(err)
	}

	// Check that the update took
	updatedFile, err := m.GetFile(fId)
	if err != nil {
		t.Fatal(err)
	}
	// Link should not have changed
	if !bytes.Equal(updatedFile.Link, testBytes) {
		t.Fatalf("Link should not have changed")
	}
	// Other attributes should have changed
	if !bytes.Equal(updatedFile.Data, newBytes) {
		t.Fatalf("Data should have updated")
	}
	if !updatedFile.Timestamp.Equal(newTs) {
		t.Fatalf("TS should have updated, expected %s got %s",
			newTs, updatedFile.Timestamp)
	}
	if updatedFile.Status != newStatus {
		t.Fatalf("Status should have updated")
	}

	// Delete the row
	err = m.DeleteFile(fId)
	if err != nil {
		t.Fatal(err)
	}

	// Check that the delete operation took and get provides the expected error
	_, err = m.GetFile(fId)
	if err == nil || !errors.Is(channels.NoMessageErr, err) {
		t.Fatal(err)
	}
}

// Happy path, insert message and look it up
func TestWasmModel_GetMessage(t *testing.T) {
	cipher, err := idbCrypto.NewCipher(
		[]byte("testPass"), []byte("testSalt"), 128, csprng.NewSystemRNG())
	if err != nil {
		t.Fatalf("Failed to create cipher")
	}
	for _, c := range []idbCrypto.Cipher{nil, cipher} {
		cs := ""
		if c != nil {
			cs = "_withCipher"
		}
		testString := "TestWasmModel_GetMessage" + cs
		t.Run(testString, func(t *testing.T) {
			storage.GetLocalStorage().Clear()
			testMsgId := message.DeriveChannelMessageID(&id.ID{1}, 0, []byte(testString))

			eventModel, err := newWASMModel(testString, c,
				&dummyCbs{})
			if err != nil {
				t.Fatal(err)
			}

			testMsg := buildMessage(id.NewIdFromBytes([]byte(testString), t).Marshal(),
				testMsgId.Bytes(), nil, testString, testString,
				[]byte{8, 6, 7, 5}, 0, 0, netTime.Now(),
				time.Second, 0, 0, false, false, channels.Sent)
			_, err = eventModel.upsertMessage(testMsg)
			if err != nil {
				t.Fatal(err)
			}

			msg, err := eventModel.GetMessage(testMsgId)
			if err != nil {
				t.Fatal(err)
			}
			if msg.UUID == 0 {
				t.Fatalf("Expected to get a UUID!")
			}
		})
	}
}

// Happy path, insert message and delete it
func TestWasmModel_DeleteMessage(t *testing.T) {
	storage.GetLocalStorage().Clear()
	testString := "TestWasmModel_DeleteMessage"
	testMsgId := message.DeriveChannelMessageID(&id.ID{1}, 0, []byte(testString))
	eventModel, err := newWASMModel(testString, nil, &dummyCbs{})
	if err != nil {
		t.Fatal(err)
	}

	// Insert a message
	testMsg := buildMessage([]byte(testString), testMsgId.Bytes(), nil,
		testString, testString, []byte{8, 6, 7, 5}, 0, 0, netTime.Now(),
		time.Second, 0, 0, false, false, channels.Sent)
	_, err = eventModel.upsertMessage(testMsg)
	if err != nil {
		t.Fatal(err)
	}

	// Check the resulting status
	results, err := impl.Dump(eventModel.db, messageStoreName)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("Expected 1 message to exist")
	}

	// Delete the message
	err = eventModel.DeleteMessage(testMsgId)
	if err != nil {
		t.Fatal(err)
	}

	// Check the resulting status
	results, err = impl.Dump(eventModel.db, messageStoreName)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Fatalf("Expected no messages to exist")
	}
}

// Test wasmModel.UpdateSentStatus happy path and ensure fields don't change.
func Test_wasmModel_UpdateSentStatus(t *testing.T) {
	cipher, err := idbCrypto.NewCipher(
		[]byte("testPass"), []byte("testSalt"), 128, csprng.NewSystemRNG())
	if err != nil {
		t.Fatalf("Failed to create cipher")
	}
	for _, c := range []idbCrypto.Cipher{nil, cipher} {
		cs := ""
		if c != nil {
			cs = "_withCipher"
		}
		t.Run("Test_wasmModel_UpdateSentStatus"+cs, func(t *testing.T) {
			storage.GetLocalStorage().Clear()
			testString := "Test_wasmModel_UpdateSentStatus" + cs
			testMsgId := message.DeriveChannelMessageID(
				&id.ID{1}, 0, []byte(testString))
			eventModel, err2 := newWASMModel(testString, c,
				&dummyCbs{})
			if err2 != nil {
				t.Fatal(err)
			}

			cid, err := id.NewRandomID(csprng.NewSystemRNG(),
				id.DummyUser.GetType())
			require.NoError(t, err)

			// Store a test message
			testMsg := buildMessage(cid.Bytes(), testMsgId.Bytes(), nil,
				testString, testString, []byte{8, 6, 7, 5}, 0, 0,
				netTime.Now(), time.Second, 0, 0, false, false, channels.Sent)
			uuid, err2 := eventModel.upsertMessage(testMsg)
			if err2 != nil {
				t.Fatal(err2)
			}

			// Ensure one message is stored
			results, err2 := impl.Dump(eventModel.db, messageStoreName)
			if err2 != nil {
				t.Fatal(err2)
			}
			if len(results) != 1 {
				t.Fatalf("Expected 1 message to exist")
			}

			// Update the sentStatus
			expectedStatus := channels.Failed
			eventModel.UpdateFromUUID(
				uuid, nil, nil, nil, nil, nil, &expectedStatus)

			// Check the resulting status
			results, err = impl.Dump(eventModel.db, messageStoreName)
			if err != nil {
				t.Fatal(err)
			}
			if len(results) != 1 {
				t.Fatalf("Expected 1 message to exist")
			}
			resultMsg := &Message{}
			err = json.Unmarshal([]byte(results[0]), resultMsg)
			if err != nil {
				t.Fatal(err)
			}
			if resultMsg.Status != uint8(expectedStatus) {
				t.Fatalf("Unexpected Status: %v", resultMsg.Status)
			}

			// Make sure other fields didn't change
			if resultMsg.Nickname != testString {
				t.Fatalf("Unexpected Nickname: %v", resultMsg.Nickname)
			}
		})
	}
}

// Smoke test wasmModel.JoinChannel/wasmModel.LeaveChannel happy paths.
func Test_wasmModel_JoinChannel_LeaveChannel(t *testing.T) {
	cipher, err := idbCrypto.NewCipher(
		[]byte("testPass"), []byte("testSalt"), 128, csprng.NewSystemRNG())
	if err != nil {
		t.Fatalf("Failed to create cipher")
	}
	for _, c := range []idbCrypto.Cipher{nil, cipher} {
		cs := ""
		if c != nil {
			cs = "_withCipher"
		}
		t.Run("Test_wasmModel_JoinChannel_LeaveChannel"+cs, func(t *testing.T) {
			storage.GetLocalStorage().Clear()
			eventModel, err2 := newWASMModel("test", c, &dummyCbs{})
			if err2 != nil {
				t.Fatal(err2)
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
			results, err2 := impl.Dump(eventModel.db, channelStoreName)
			if err2 != nil {
				t.Fatal(err2)
			}
			if len(results) != 2 {
				t.Fatalf("Expected 2 channels to exist")
			}
			eventModel.LeaveChannel(testChannel.ReceptionID)
			results, err = impl.Dump(eventModel.db, channelStoreName)
			if err != nil {
				t.Fatal(err)
			}
			if len(results) != 1 {
				t.Fatalf("Expected 1 channels to exist")
			}
		})
	}
}

// Test UUID gets returned when different messages are added.
func Test_wasmModel_UUIDTest(t *testing.T) {
	cipher, err := idbCrypto.NewCipher(
		[]byte("testPass"), []byte("testSalt"), 128, csprng.NewSystemRNG())
	if err != nil {
		t.Fatalf("Failed to create cipher")
	}
	for _, c := range []idbCrypto.Cipher{nil, cipher} {
		cs := ""
		if c != nil {
			cs = "_withCipher"
		}
		t.Run("Test_wasmModel_UUIDTest"+cs, func(t *testing.T) {
			storage.GetLocalStorage().Clear()
			testString := "testHello" + cs
			eventModel, err2 := newWASMModel(testString, c,
				&dummyCbs{})
			if err2 != nil {
				t.Fatal(err2)
			}

			uuids := make([]uint64, 10)

			for i := 0; i < 10; i++ {
				// Store a test message
				channelID := id.NewIdFromBytes([]byte(testString), t)
				msgID := message.ID{}
				copy(msgID[:], testString+fmt.Sprintf("%d", i))
				rnd := rounds.Round{ID: id.Round(42)}
				uuid := eventModel.ReceiveMessage(channelID, msgID, "test",
					testString+fmt.Sprintf("%d", i), []byte{8, 6, 7, 5}, 0, 0,
					netTime.Now(), time.Hour, rnd, 0, channels.Sent, false)
				uuids[i] = uuid
			}

			for i := 0; i < 10; i++ {
				for j := i + 1; j < 10; j++ {
					if uuids[i] == uuids[j] {
						t.Fatalf("uuid failed: %d[%d] == %d[%d]",
							uuids[i], i, uuids[j], j)
					}
				}
			}
		})
	}
}

// Tests if the same message ID being sent always returns the same UUID.
func Test_wasmModel_DuplicateReceives(t *testing.T) {
	cipher, err := idbCrypto.NewCipher(
		[]byte("testPass"), []byte("testSalt"), 128, csprng.NewSystemRNG())
	if err != nil {
		t.Fatalf("Failed to create cipher")
	}
	for _, c := range []idbCrypto.Cipher{nil, cipher} {
		cs := ""
		if c != nil {
			cs = "_withCipher"
		}
		testString := "Test_wasmModel_DuplicateReceives" + cs
		t.Run(testString, func(t *testing.T) {
			storage.GetLocalStorage().Clear()
			eventModel, err := newWASMModel(testString, c,
				&dummyCbs{})
			if err != nil {
				t.Fatal(err)
			}

			// Store a test message
			msgID := message.ID{}
			copy(msgID[:], testString)
			channelID := id.NewIdFromBytes([]byte(testString), t)
			rnd := rounds.Round{ID: id.Round(42)}
			uuid := eventModel.ReceiveMessage(channelID, msgID, "test",
				testString, []byte{8, 6, 7, 5}, 0, 0,
				netTime.Now(), time.Hour, rnd, 0, channels.Sent, false)
			if uuid != 1 {
				t.Fatalf("Expected UUID to be one for first receive")
			}

			// Store duplicate messages with same messageID
			referenceID := uint64(0)
			for i := 0; i < 10; i++ {
				uuid = eventModel.ReceiveMessage(channelID, msgID, "test",
					testString+fmt.Sprintf("%d", i), []byte{8, 6, 7, 5}, 0, 0,
					netTime.Now(), time.Hour, rnd, 0, channels.Sent, false)
				if referenceID == 0 {
					referenceID = uuid
				}
				require.Equal(t, referenceID, uuid,
					"UUID must be identical for duplicate receives")
			}
		})
	}
}

// Happy path: Inserts many messages, deletes some, and checks that the final
// result is as expected.
func Test_wasmModel_deleteMsgByChannel(t *testing.T) {
	cipher, err := idbCrypto.NewCipher(
		[]byte("testPass"), []byte("testSalt"), 128, csprng.NewSystemRNG())
	if err != nil {
		t.Fatalf("Failed to create cipher")
	}
	for _, c := range []idbCrypto.Cipher{nil, cipher} {
		cs := ""
		if c != nil {
			cs = "_withCipher"
		}
		testString := "Test_wasmModel_deleteMsgByChannel" + cs
		t.Run(testString, func(t *testing.T) {
			storage.GetLocalStorage().Clear()
			totalMessages := 10
			expectedMessages := 5
			eventModel, err := newWASMModel(testString, c,
				&dummyCbs{})
			if err != nil {
				t.Fatal(err)
			}

			// Create a test channel id
			deleteChannel := id.NewIdFromString("deleteMe", id.Generic, t)
			keepChannel := id.NewIdFromString("dontDeleteMe", id.Generic, t)

			// Store some test messages
			for i := 0; i < totalMessages; i++ {
				testStr := testString + strconv.Itoa(i)

				// Interleave the channel id to ensure cursor is behaving intelligently
				thisChannel := deleteChannel
				if i%2 == 0 {
					thisChannel = keepChannel
				}

				testMsgId := message.DeriveChannelMessageID(
					&id.ID{byte(i)}, 0, []byte(testStr))
				eventModel.ReceiveMessage(thisChannel, testMsgId, testStr,
					testStr, []byte{8, 6, 7, 5}, 0, 0, netTime.Now(),
					time.Second, rounds.Round{ID: id.Round(0)}, 0,
					channels.Sent, false)
			}

			// Check pre-results
			result, err := impl.Dump(eventModel.db, messageStoreName)
			if err != nil {
				t.Fatal(err)
			}
			if len(result) != totalMessages {
				t.Fatalf("Expected %d messages, got %d", totalMessages, len(result))
			}

			// Do delete
			err = eventModel.deleteMsgByChannel(deleteChannel)
			if err != nil {
				t.Error(err)
			}

			// Check final results
			result, err = impl.Dump(eventModel.db, messageStoreName)
			if err != nil {
				t.Fatal(err)
			}
			if len(result) != expectedMessages {
				t.Fatalf("Expected %d messages, got %d", expectedMessages, len(result))
			}
		})
	}
}

// This test is designed to prove the behavior of unique indexes.
// Inserts will not fail, they simply will not happen.
func TestWasmModel_receiveHelper_UniqueIndex(t *testing.T) {
	cipher, err := idbCrypto.NewCipher(
		[]byte("testPass"), []byte("testSalt"), 128, csprng.NewSystemRNG())
	if err != nil {
		t.Fatalf("Failed to create cipher")
	}
	for i, c := range []idbCrypto.Cipher{nil, cipher} {
		cs := ""
		if c != nil {
			cs = "_withCipher"
		}
		t.Run("TestWasmModel_receiveHelper_UniqueIndex"+cs, func(t *testing.T) {
			storage.GetLocalStorage().Clear()
			testString := fmt.Sprintf("test_receiveHelper_UniqueIndex_%d", i)
			eventModel, err := newWASMModel(testString, c,
				&dummyCbs{})
			if err != nil {
				t.Fatal(err)
			}

			// Ensure index is unique
			txn, err := eventModel.db.Transaction(
				idb.TransactionReadOnly, messageStoreName)
			if err != nil {
				t.Fatal(err)
			}
			store, err := txn.ObjectStore(messageStoreName)
			if err != nil {
				t.Fatal(err)
			}
			idx, err := store.Index(messageStoreMessageIndex)
			if err != nil {
				t.Fatal(err)
			}
			if isUnique, err3 := idx.Unique(); !isUnique {
				t.Fatalf("Index is not unique!")
			} else if err3 != nil {
				t.Fatal(err3)
			}

			testMsgId := message.DeriveChannelMessageID(&id.ID{1}, 0, []byte(testString))
			testMsg := buildMessage([]byte(testString), testMsgId.Bytes(), nil,
				testString, testString, []byte{8, 6, 7, 5}, 0, 0,
				netTime.Now(), time.Second, 0, 0, false, false, channels.Sent)

			testMsgId2 := message.DeriveChannelMessageID(&id.ID{2}, 0, []byte(testString))
			testMsg2 := buildMessage([]byte(testString), testMsgId2.Bytes(), nil,
				testString, testString, []byte{8, 6, 7, 5}, 0, 0,
				netTime.Now(), time.Second, 0, 0, false, false, channels.Sent)

			// First message insert should succeed
			uuid, err := eventModel.upsertMessage(testMsg)
			if err != nil {
				t.Fatal(err)
			}

			// The duplicate entry should fail
			duplicateUuid, err := eventModel.upsertMessage(testMsg)
			if err == nil {
				t.Fatal("Expected error to happen")
			}
			if duplicateUuid != 0 {
				t.Fatalf("Expected UUID %d to be 0", duplicateUuid)
			}

			// Now insert a message with a different message ID from the first
			uuid2, err := eventModel.upsertMessage(testMsg2)
			if err != nil {
				t.Fatal(err)
			}
			if uuid2 == uuid {
				t.Fatalf("Expected UUID %d to NOT match %d", uuid, uuid2)
			}

			// Except this time, we update the second entry to have the same
			// message ID as the first
			testMsg2.MessageID = testMsgId.Bytes()
			duplicateUuid, err = eventModel.upsertMessage(testMsg)
			if err == nil {
				t.Fatal("Expected error to happen")
			}
			if duplicateUuid != 0 {
				t.Fatalf("Expected UUID %d to be 0", uuid)
			}
		})
	}
}

func TestUpsertMessageDuplicates(t *testing.T) {
	msgJSON := []byte(`{"pubkey":"4gFD38jRhf/aNDRzaguLJHHgTncs0pwvNiRvjTBOaXk=",
                "dm_token":3173028031,
		"nickname":"",
		"message_id":"BTi+V2anHhrRAEcCMm6g2Bs9oKn7PhQzCzvM7x8Ml7A=",
		"channel_id":"Gt1T8t1GcnQXijiOjX5yjiLt3F8tnPB9S70ntBx86h8D",
		"parent_message_id":null,
		"pinned":false,
		"hidden":false,
		"round":151369004,
		"codeset_version":0,
		"timestamp": "2023-06-06T19:39:27.281204113Z",
		"lease_v2":"30000000000",
		"status":2,
		"type":1,
		"text":"KzT25s76K9AEIv0HadhjULkU37dkCUp/8N9+5UkkA8pLEBK71P7KFMr2xii/enYr4jpQcayNm82WD4wisimkK2wtaKb0hJIELAxJHxHilF2qVcDYn7zmx2ZiYkerZGTlf7DkHPZrA+EJJIu9gZ4ApR8TNVf15GxPSDGKHV6iv7vEMrrdoi4JKOFTpNsUjjbDykc91ADmmRlh5LUR/lqyt7bEasyVY2Zo+UiZpJrUJII0fhhn0HLbfs6ekg1tkq+vbPap3vNoebLZag3nEnMk+JqNjVW93BFICSMFv8cYX9q0sSK6sgHtz0HX9b7xTEFVHlCb2Zv/ErOX2LiedsgM4aGpOvjGYQM9ludCutVJArfM/5ejupMpvVPimY69QdevqxMm6v7c8H3abLQf4iKNVoaTaRygYFYArTWWoKep7BDZhjhCd4XnILGLjq7oeBfb3enMUSmeYabIuQpjAfRXU6cZh4cVomhEbkF5kj/rAc3GLzWtN1XcE5PC8bTL3m6YVN6hMCd1S7dsnDQ4FJQoDoqYzmTLD7jtyQujS86JsPkkt7kAIi14VxHTNEAM07cN6rH4VifWWMgQo4djPJ+nnsW9zUnylT9KS1GGyeaUhGC7GmrJGd7DmfHrhdWb9iMboKyH7VTm5C6X71JW1AhmdFfjtIHmWfhbAFsrP2gwkb4e+F86urUmK5LTT6qhGklUsXsmMFjwiEUCBWLSOZLKfITkNOBGIuGGcMi8Jtvwo1f4DHQ6sm/Y8qaJFmesyVKpA3MAVNawqlZJSJXflUA/mpJprFxLm/J3mdnxkLrqQSeSCAxHgpEdH3jiSPNhmn6QaJOVT+wotm+8lAq7qWgBWfMmcVSgWdAfVwujeP71Uo1y24/Z4+jXulFz1hPfbfmHk6/Qz4JlSgHZIqgOS1FTsmXFvBUe4moHR/QFDEXnz2LroRIavt9gWZPE3z1FIcjwSmtUzvpE8gZud+dKVVrjm8YPseq+IJ/AwRRlkUR5D0vf1TwEm0yQ7E+EotFvibQ="
	}`)

	// Initial call
	msg1 := &Message{}
	err := json.Unmarshal(msgJSON, msg1)
	require.NoError(t, err)
	msg1.Status = 0

	// Second call
	msg2 := &Message{}
	err = json.Unmarshal(msgJSON, msg2)
	require.NoError(t, err)
	msg2.Status = 1

	// Final call
	msg3 := &Message{}
	err = json.Unmarshal(msgJSON, msg3)
	require.NoError(t, err)
	msg3.Status = 2

	cipher, err := idbCrypto.NewCipher(
		[]byte("testPass"), []byte("testSalt"), 128,
		csprng.NewSystemRNG())
	require.NoError(t, err)
	testString := "test_duplicateUpsertMessage"
	eventModel, err := newWASMModel(testString, cipher,
		&dummyCbs{})
	require.NoError(t, err)

	uuid, err := eventModel.upsertMessage(msg1)
	require.NoError(t, err)
	require.NotEqual(t, uuid, 0)

	uuid2, err := eventModel.upsertMessage(msg2)
	require.NoError(t, err)
	require.Equal(t, uuid, uuid2)

	uuid3, err := eventModel.upsertMessage(msg3)
	require.NoError(t, err)
	require.Equal(t, uuid, uuid3)

	msgID, err := message.UnmarshalID(msg1.MessageID)
	require.NoError(t, err)

	modelMsg, err := eventModel.GetMessage(msgID)
	require.NoError(t, err)
	require.Equal(t, channels.Delivered, modelMsg.Status)
}
