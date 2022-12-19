////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/client/v4/auth"
	"gitlab.com/elixxir/client/v4/catalog"
	"gitlab.com/elixxir/client/v4/channels"
	"gitlab.com/elixxir/client/v4/cmix"
	"gitlab.com/elixxir/client/v4/cmix/message"
	"gitlab.com/elixxir/client/v4/connect"
	"gitlab.com/elixxir/client/v4/e2e/ratchet/partner"
	ftE2e "gitlab.com/elixxir/client/v4/fileTransfer/e2e"
	"gitlab.com/elixxir/client/v4/groupChat/groupStore"
	"gitlab.com/elixxir/client/v4/restlike"
	"gitlab.com/elixxir/client/v4/single"
	"gitlab.com/elixxir/crypto/broadcast"
	"gitlab.com/elixxir/crypto/contact"
	"gitlab.com/elixxir/crypto/cyclic"
	"gitlab.com/elixxir/crypto/fileTransfer"
	"gitlab.com/elixxir/crypto/group"
	cryptoMessage "gitlab.com/elixxir/crypto/message"
	"gitlab.com/elixxir/primitives/fact"
	"gitlab.com/elixxir/primitives/format"
	"gitlab.com/xx_network/primitives/id"
	"gitlab.com/xx_network/primitives/id/ephemeral"
	"gitlab.com/xx_network/primitives/ndf"
	"gitlab.com/xx_network/primitives/netTime"
)

// These objects are imported so that doc linking on pkg.go.dev does not require
// the entire package URL.
var (
	_ = id.ID{}
	_ = ephemeral.Id{}
	_ = id.Round(0)
	_ = contact.Contact{}
	_ = cmix.RoundResult{}
	_ = single.RequestParams{}
	_ = channels.Manager(nil)
	_ = catalog.MessageType(0)
	_ = connect.Callback(nil)
	_ = partner.Manager(nil)
	_ = ndf.NetworkDefinition{}
	_ = cryptoMessage.ID{}
	_ = channels.SentStatus(0)
	_ = ftE2e.Params{}
	_ = fileTransfer.TransferID{}
	_ = message.ServiceList{}
	_ = group.Membership{}
	_ = groupStore.Group{}
	_ = cyclic.Int{}
	_ = fact.FactList{}
	_ = fact.Fact{}
	_ = jwalterweatherman.LogListener(nil)
	_ = format.Message{}
	_ = restlike.Message{}
	_ = auth.Callbacks(nil)
	_ = broadcast.PrivacyLevel(0)
	_ = broadcast.Channel{}
	_ = netTime.Now
)
