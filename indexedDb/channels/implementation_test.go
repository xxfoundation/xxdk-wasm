////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package channels

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"github.com/hack-pad/go-indexeddb/idb"
	"gitlab.com/elixxir/crypto/message"
	"gitlab.com/elixxir/xxdk-wasm/indexedDb"
	"gitlab.com/elixxir/xxdk-wasm/storage"
	"gitlab.com/xx_network/crypto/csprng"
	"gitlab.com/xx_network/primitives/netTime"
	"os"
	"strconv"
	"testing"
	"time"

	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/client/v4/channels"
	"gitlab.com/elixxir/client/v4/cmix/rounds"
	cryptoBroadcast "gitlab.com/elixxir/crypto/broadcast"
	cryptoChannel "gitlab.com/elixxir/crypto/channel"
	"gitlab.com/xx_network/primitives/id"
)

func TestMain(m *testing.M) {
	jww.SetStdoutThreshold(jww.LevelDebug)
	os.Exit(m.Run())
}

func dummyReceivedMessageCB(uint64, *id.ID, bool)      {}
func dummyDeletedMessageCB(message.ID)                 {}
func dummyMutedUserCB(*id.ID, ed25519.PublicKey, bool) {}

// Happy path, insert message and look it up
func TestWasmModel_msgIDLookup(t *testing.T) {
	cipher, err := cryptoChannel.NewCipher(
		[]byte("testPass"), []byte("testSalt"), 128, csprng.NewSystemRNG())
	if err != nil {
		t.Fatalf("Failed to create cipher")
	}
	for _, c := range []cryptoChannel.Cipher{nil, cipher} {
		cs := ""
		if c != nil {
			cs = "_withCipher"
		}
		t.Run(fmt.Sprintf("TestWasmModel_msgIDLookup%s", cs), func(t *testing.T) {

			storage.GetLocalStorage().Clear()
			testString := "TestWasmModel_msgIDLookup" + cs
			testMsgId := message.DeriveChannelMessageID(&id.ID{1}, 0, []byte(testString))

			eventModel, err2 := newWASMModel(testString, c,
				dummyReceivedMessageCB, dummyDeletedMessageCB, dummyMutedUserCB)
			if err2 != nil {
				t.Fatalf("%+v", err2)
			}

			testMsg := buildMessage([]byte(testString), testMsgId.Bytes(), nil,
				testString, []byte(testString), []byte{8, 6, 7, 5}, 0, 0,
				netTime.Now(), time.Second, 0, 0, false, false, channels.Sent)
			_, err = eventModel.receiveHelper(testMsg, false)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			msg, err2 := eventModel.msgIDLookup(testMsgId)
			if err2 != nil {
				t.Fatalf("%+v", err2)
			}
			if msg.ID == 0 {
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
	eventModel, err := newWASMModel(testString, nil,
		dummyReceivedMessageCB, dummyDeletedMessageCB, dummyMutedUserCB)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	// Insert a message
	testMsg := buildMessage([]byte(testString), testMsgId.Bytes(), nil,
		testString, []byte(testString), []byte{8, 6, 7, 5}, 0, 0, netTime.Now(),
		time.Second, 0, 0, false, false, channels.Sent)
	_, err = eventModel.receiveHelper(testMsg, false)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	// Check the resulting status
	results, err := indexedDb.Dump(eventModel.db, messageStoreName)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	if len(results) != 1 {
		t.Fatalf("Expected 1 message to exist")
	}

	// Delete the message
	err = eventModel.DeleteMessage(testMsgId)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	// Check the resulting status
	results, err = indexedDb.Dump(eventModel.db, messageStoreName)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	if len(results) != 0 {
		t.Fatalf("Expected no messages to exist")
	}
}

// Test wasmModel.UpdateSentStatus happy path and ensure fields don't change.
func Test_wasmModel_UpdateSentStatus(t *testing.T) {
	cipher, err := cryptoChannel.NewCipher(
		[]byte("testPass"), []byte("testSalt"), 128, csprng.NewSystemRNG())
	if err != nil {
		t.Fatalf("Failed to create cipher")
	}
	for _, c := range []cryptoChannel.Cipher{nil, cipher} {
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
				dummyReceivedMessageCB, dummyDeletedMessageCB, dummyMutedUserCB)
			if err2 != nil {
				t.Fatalf("%+v", err2)
			}

			// Store a test message
			testMsg := buildMessage([]byte(testString), testMsgId.Bytes(), nil,
				testString, []byte(testString), []byte{8, 6, 7, 5}, 0, 0, netTime.Now(),
				time.Second, 0, 0, false, false, channels.Sent)
			uuid, err := eventModel.receiveHelper(testMsg, false)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			// Ensure one message is stored
			results, err := indexedDb.Dump(eventModel.db, messageStoreName)
			if err != nil {
				t.Fatalf("%+v", err)
			}
			if len(results) != 1 {
				t.Fatalf("Expected 1 message to exist")
			}

			// Update the sentStatus
			expectedStatus := channels.Failed
			eventModel.UpdateFromUUID(
				uuid, nil, nil, nil, nil, nil, &expectedStatus)

			// Check the resulting status
			results, err2 = indexedDb.Dump(eventModel.db, messageStoreName)
			if err2 != nil {
				t.Fatalf("%+v", err2)
			}
			if len(results) != 1 {
				t.Fatalf("Expected 1 message to exist")
			}
			resultMsg := &Message{}
			err2 = json.Unmarshal([]byte(results[0]), resultMsg)
			if err2 != nil {
				t.Fatalf("%+v", err2)
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
	cipher, err := cryptoChannel.NewCipher(
		[]byte("testPass"), []byte("testSalt"), 128, csprng.NewSystemRNG())
	if err != nil {
		t.Fatalf("Failed to create cipher")
	}
	for _, c := range []cryptoChannel.Cipher{nil, cipher} {
		cs := ""
		if c != nil {
			cs = "_withCipher"
		}
		t.Run("Test_wasmModel_JoinChannel_LeaveChannel"+cs, func(t *testing.T) {
			storage.GetLocalStorage().Clear()
			eventModel, err2 := newWASMModel("test", c,
				dummyReceivedMessageCB, dummyDeletedMessageCB, dummyMutedUserCB)
			if err2 != nil {
				t.Fatalf("%+v", err2)
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
			results, err2 := indexedDb.Dump(eventModel.db, channelsStoreName)
			if err2 != nil {
				t.Fatalf("%+v", err2)
			}
			if len(results) != 2 {
				t.Fatalf("Expected 2 channels to exist")
			}
			eventModel.LeaveChannel(testChannel.ReceptionID)
			results, err2 = indexedDb.Dump(eventModel.db, channelsStoreName)
			if err2 != nil {
				t.Fatalf("%+v", err2)
			}
			if len(results) != 1 {
				t.Fatalf("Expected 1 channels to exist")
			}
		})
	}
}

// Test UUID gets returned when different messages are added.
func Test_wasmModel_UUIDTest(t *testing.T) {
	cipher, err := cryptoChannel.NewCipher(
		[]byte("testPass"), []byte("testSalt"), 128, csprng.NewSystemRNG())
	if err != nil {
		t.Fatalf("Failed to create cipher")
	}
	for _, c := range []cryptoChannel.Cipher{nil, cipher} {
		cs := ""
		if c != nil {
			cs = "_withCipher"
		}
		t.Run("Test_wasmModel_UUIDTest"+cs, func(t *testing.T) {
			storage.GetLocalStorage().Clear()
			testString := "testHello" + cs
			eventModel, err2 := newWASMModel(testString, c,
				dummyReceivedMessageCB, dummyDeletedMessageCB, dummyMutedUserCB)
			if err2 != nil {
				t.Fatalf("%+v", err2)
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
	cipher, err := cryptoChannel.NewCipher(
		[]byte("testPass"), []byte("testSalt"), 128, csprng.NewSystemRNG())
	if err != nil {
		t.Fatalf("Failed to create cipher")
	}
	for _, c := range []cryptoChannel.Cipher{nil, cipher} {
		cs := ""
		if c != nil {
			cs = "_withCipher"
		}
		t.Run("Test_wasmModel_DuplicateReceives"+cs, func(t *testing.T) {
			storage.GetLocalStorage().Clear()
			testString := "testHello"
			eventModel, err2 := newWASMModel(testString, c,
				dummyReceivedMessageCB, dummyDeletedMessageCB, dummyMutedUserCB)
			if err2 != nil {
				t.Fatalf("%+v", err2)
			}

			uuids := make([]uint64, 10)

			msgID := message.ID{}
			copy(msgID[:], testString)
			for i := 0; i < 10; i++ {
				// Store a test message
				channelID := id.NewIdFromBytes([]byte(testString), t)
				rnd := rounds.Round{ID: id.Round(42)}
				uuid := eventModel.ReceiveMessage(channelID, msgID, "test",
					testString+fmt.Sprintf("%d", i), []byte{8, 6, 7, 5}, 0, 0,
					netTime.Now(), time.Hour, rnd, 0, channels.Sent, false)
				uuids[i] = uuid
			}

			for i := 0; i < 10; i++ {
				for j := i + 1; j < 10; j++ {
					if uuids[i] != uuids[j] {
						t.Fatalf("uuid failed: %d[%d] != %d[%d]",
							uuids[i], i, uuids[j], j)
					}
				}
			}
		})
	}
}

// Happy path: Inserts many messages, deletes some, and checks that the final
// result is as expected.
func Test_wasmModel_deleteMsgByChannel(t *testing.T) {
	cipher, err := cryptoChannel.NewCipher(
		[]byte("testPass"), []byte("testSalt"), 128, csprng.NewSystemRNG())
	if err != nil {
		t.Fatalf("Failed to create cipher")
	}
	for _, c := range []cryptoChannel.Cipher{nil, cipher} {
		cs := ""
		if c != nil {
			cs = "_withCipher"
		}
		t.Run("Test_wasmModel_deleteMsgByChannel"+cs, func(t *testing.T) {
			storage.GetLocalStorage().Clear()
			testString := "test_deleteMsgByChannel"
			totalMessages := 10
			expectedMessages := 5
			eventModel, err2 := newWASMModel(testString, c,
				dummyReceivedMessageCB, dummyDeletedMessageCB, dummyMutedUserCB)
			if err2 != nil {
				t.Fatalf("%+v", err2)
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
			result, err2 := indexedDb.Dump(eventModel.db, messageStoreName)
			if err2 != nil {
				t.Fatalf("%+v", err2)
			}
			if len(result) != totalMessages {
				t.Errorf("Expected %d messages, got %d", totalMessages, len(result))
			}

			// Do delete
			err = eventModel.deleteMsgByChannel(deleteChannel)
			if err != nil {
				t.Error(err)
			}

			// Check final results
			result, err = indexedDb.Dump(eventModel.db, messageStoreName)
			if err != nil {
				t.Fatalf("%+v", err)
			}
			if len(result) != expectedMessages {
				t.Errorf("Expected %d messages, got %d", expectedMessages, len(result))
			}
		})
	}
}

// This test is designed to prove the behavior of unique indexes.
// Inserts will not fail, they simply will not happen.
func TestWasmModel_receiveHelper_UniqueIndex(t *testing.T) {
	cipher, err := cryptoChannel.NewCipher(
		[]byte("testPass"), []byte("testSalt"), 128, csprng.NewSystemRNG())
	if err != nil {
		t.Fatalf("Failed to create cipher")
	}
	for i, c := range []cryptoChannel.Cipher{nil, cipher} {
		cs := ""
		if c != nil {
			cs = "_withCipher"
		}
		t.Run("TestWasmModel_receiveHelper_UniqueIndex"+cs, func(t *testing.T) {
			storage.GetLocalStorage().Clear()
			testString := fmt.Sprintf("test_receiveHelper_UniqueIndex_%d", i)
			eventModel, err2 := newWASMModel(testString, c,
				dummyReceivedMessageCB, dummyDeletedMessageCB, dummyMutedUserCB)
			if err2 != nil {
				t.Fatal(err2)
			}

			// Ensure index is unique
			txn, err2 := eventModel.db.Transaction(
				idb.TransactionReadOnly, messageStoreName)
			if err2 != nil {
				t.Fatal(err2)
			}
			store, err2 := txn.ObjectStore(messageStoreName)
			if err2 != nil {
				t.Fatal(err2)
			}
			idx, err2 := store.Index(messageStoreMessageIndex)
			if err2 != nil {
				t.Fatal(err2)
			}
			if isUnique, err3 := idx.Unique(); !isUnique {
				t.Fatalf("Index is not unique!")
			} else if err3 != nil {
				t.Fatal(err3)
			}

			// First message insert should succeed
			testMsgId := message.DeriveChannelMessageID(&id.ID{1}, 0, []byte(testString))
			testMsg := buildMessage([]byte(testString), testMsgId.Bytes(), nil,
				testString, []byte(testString), []byte{8, 6, 7, 5}, 0, 0,
				netTime.Now(), time.Second, 0, 0, false, false, channels.Sent)
			uuid, err := eventModel.receiveHelper(testMsg, false)
			if err != nil {
				t.Fatal(err)
			}

			// The duplicate entry should return the same UUID
			duplicateUuid, err := eventModel.receiveHelper(testMsg, false)
			if err != nil {
				t.Fatal(err)
			}
			if uuid != duplicateUuid {
				t.Fatalf("Expected UUID %d to match %d", uuid, duplicateUuid)
			}

			// Now insert a message with a different message ID from the first
			testMsgId2 := message.DeriveChannelMessageID(
				&id.ID{2}, 0, []byte(testString))
			testMsg = buildMessage([]byte(testString), testMsgId2.Bytes(), nil,
				testString, []byte(testString), []byte{8, 6, 7, 5}, 0, 0,
				netTime.Now(), time.Second, 0, 0, false, false, channels.Sent)
			uuid2, err := eventModel.receiveHelper(testMsg, false)
			if err != nil {
				t.Fatal(err)
			}
			if uuid2 == uuid {
				t.Fatalf("Expected UUID %d to NOT match %d", uuid, duplicateUuid)
			}

			// Except this time, we update the second entry to have the same
			// message ID as the first
			testMsg.ID = uuid
			testMsg.MessageID = testMsgId.Bytes()
			duplicateUuid2, err := eventModel.receiveHelper(testMsg, true)
			if err != nil {
				t.Fatal(err)
			}
			if duplicateUuid2 != duplicateUuid {
				t.Fatalf("Expected UUID %d to match %d", uuid, duplicateUuid)
			}
		})
	}
}
