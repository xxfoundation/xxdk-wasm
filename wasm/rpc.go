////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"syscall/js"

	"gitlab.com/elixxir/client/v4/bindings"
	"gitlab.com/elixxir/wasm-utils/utils"

	jww "github.com/spf13/jwalterweatherman"
)

// Client RPC Interface

// RPCSend sends an RPC request and returns the results.
//
// Parameters:
//   - args[0] - ID of [Cmix] object in tracker (int). This can be retrieved
//     using [Cmix.GetID].
//   - args[1] - Bytes of a public reception ID to contact the server. This is
//     a [Cmix.ReceptionID] of 33 bytes (Uint8Array).
//   - args[2] - Bytes of a public key of the RPC Server (Uint8Array).
//   - args[3] - Bytes of your request (Uint8Array).
//   - args[4] - Optional. A Javascript callback function to return
//     intermediate results for when the message is request is
//     sent/processed and the response. If left undefined, this prints
//     to the console log. ((json: Uint8Array) => void).
//
// Returns:
//   - Javascript representation of the [DMClient] object.
//   - Throws an error if creating the manager fails.
func RPCSend(_ js.Value, args []js.Value) any {
	cMixID := args[0].Int()
	recipient := utils.CopyBytesToGo(args[1])
	pubkey := utils.CopyBytesToGo(args[2])
	request := utils.CopyBytesToGo(args[3])
	var cb js.Value
	hasCb := false
	if len(args) >= 4 {
		cb = args[4]
		hasCb = true
	}

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		r := bindings.RPCSend(cMixID, recipient, pubkey, request)
		cbs := &rpcResponse{
			responseFn: func(response []byte) {
				if hasCb {
					cb.Invoke(utils.CopyBytesToJS(response))
					return
				}
				jww.INFO.Printf("[RPCSend] Callback: %s",
					string(response))

			},
			errFn: func(err []byte) {
				reject(utils.CopyBytesToJS(err))
			},
		}
		r.Callback(cbs)
		// We resolve only once per promise rules, which means we take
		// the final return value.
		resolve(utils.CopyBytesToJS(r.Await()))
	}
	return utils.CreatePromise(promiseFn)
}

type rpcResponse struct {
	responseFn func(response []byte)
	errFn      func(err []byte)
}

func (r *rpcResponse) Response(response []byte) {
	r.responseFn(response)
}
func (r *rpcResponse) Error(errorStr string) {
	r.errFn([]byte(errorStr))
}

// NOTE: The server RPC interface is not exposed on the WASM interface
// at this time, since this code is really intended to target
// browser-apps. If you need it, we can implement it for your pretty
// quickly. Please message partners at xx dot network via e-mail.
