////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file                                                               //
////////////////////////////////////////////////////////////////////////////////

package indexedDb

import (
	jww "github.com/spf13/jwalterweatherman"
	cryptoBroadcast "gitlab.com/elixxir/crypto/broadcast"
	"gitlab.com/xx_network/primitives/id"
	"testing"
)

func TestWasmModel_JoinChannel_LeaveChannel(t *testing.T) {
	testDbName := "test"
	jww.SetStdoutThreshold(jww.LevelTrace)

	eventModel, err := newWasmModel(testDbName)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	testChannel := &cryptoBroadcast.Channel{
		ReceptionID: id.NewIdFromString("test", id.Generic, t),
		Name:        "test",
		Description: "test",
		Salt:        nil,
		RsaPubKey:   nil,
	}
	testChannel2 := &cryptoBroadcast.Channel{
		ReceptionID: id.NewIdFromString("test2", id.Generic, t),
		Name:        "test2",
		Description: "test2",
		Salt:        nil,
		RsaPubKey:   nil,
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
