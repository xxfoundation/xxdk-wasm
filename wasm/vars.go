////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package wasm

import (
	"github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/client/catalog"
	"gitlab.com/elixxir/client/channels"
	"gitlab.com/elixxir/client/cmix"
	"gitlab.com/elixxir/client/cmix/message"
	"gitlab.com/elixxir/client/connect"
	"gitlab.com/elixxir/client/e2e/ratchet/partner"
	ftE2e "gitlab.com/elixxir/client/fileTransfer/e2e"
	"gitlab.com/elixxir/client/groupChat/groupStore"
	"gitlab.com/elixxir/client/restlike"
	"gitlab.com/elixxir/client/single"
	"gitlab.com/elixxir/crypto/channel"
	"gitlab.com/elixxir/crypto/contact"
	"gitlab.com/elixxir/crypto/cyclic"
	"gitlab.com/elixxir/crypto/fileTransfer"
	"gitlab.com/elixxir/crypto/group"
	"gitlab.com/elixxir/primitives/fact"
	"gitlab.com/elixxir/primitives/format"
	"gitlab.com/xx_network/primitives/id"
	"gitlab.com/xx_network/primitives/id/ephemeral"
	"gitlab.com/xx_network/primitives/ndf"
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
	_ = channel.MessageID{}
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
)
