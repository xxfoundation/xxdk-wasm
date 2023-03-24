////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"gitlab.com/elixxir/xxdk-wasm/src/api/utils"
	"syscall/js"
)

// GetReceptionID returns the marshalled default IDs.
//
// Returns:
//   - Marshalled bytes of [id.ID] (Uint8Array).
func (e *E2e) GetReceptionID(js.Value, []js.Value) any {
	return utils.CopyBytesToJS(e.api.GetReceptionID())
}

// DeleteContact removes a partner from [E2e]'s storage.
//
// Parameters:
//   - args[0] - Marshalled bytes of the partner [id.ID] (Uint8Array).
//
// Returns:
//   - Throws TypeError if deleting the partner fails.
func (e *E2e) DeleteContact(_ js.Value, args []js.Value) any {
	err := e.api.DeleteContact(utils.CopyBytesToGo(args[0]))
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}
	return nil
}

// GetAllPartnerIDs returns a list of all partner IDs that the user has an E2E
// relationship with.
//
// Returns:
//   - JSON of array of [id.ID] (Uint8Array).
//   - Throws TypeError if getting partner IDs fails.
func (e *E2e) GetAllPartnerIDs(js.Value, []js.Value) any {
	partnerIDs, err := e.api.GetAllPartnerIDs()
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}
	return utils.CopyBytesToJS(partnerIDs)
}

// PayloadSize returns the max payload size for a partitionable E2E message.
//
// Returns:
//   - Max payload size (int).
func (e *E2e) PayloadSize(js.Value, []js.Value) any {
	return e.api.PayloadSize()
}

// SecondPartitionSize returns the max partition payload size for all payloads
// after the first payload.
//
// Returns:
//   - Max payload size (int).
func (e *E2e) SecondPartitionSize(js.Value, []js.Value) any {
	return e.api.SecondPartitionSize()
}

// PartitionSize returns the partition payload size for the given payload index.
// The first payload is index 0.
//
// Parameters:
//   - args[0] - Payload index (int).
//
// Returns:
//   - Partition payload size (int).
func (e *E2e) PartitionSize(_ js.Value, args []js.Value) any {
	return e.api.PartitionSize(args[0].Int())
}

// FirstPartitionSize returns the max partition payload size for the first
// payload.
//
// Returns:
//   - Max partition payload size (int).
func (e *E2e) FirstPartitionSize(js.Value, []js.Value) any {
	return e.api.FirstPartitionSize()
}

// GetHistoricalDHPrivkey returns the user's marshalled historical DH private
// key.
//
// Returns:
//   - JSON of [cyclic.Int] (Uint8Array).
//   - Throws TypeError if getting the key fails.
func (e *E2e) GetHistoricalDHPrivkey(js.Value, []js.Value) any {
	privKey, err := e.api.GetHistoricalDHPrivkey()
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}
	return utils.CopyBytesToJS(privKey)
}

// GetHistoricalDHPubkey returns the user's marshalled historical DH public key.
// key.
//
// Returns:
//   - JSON of [cyclic.Int] (Uint8Array).
//   - Throws TypeError if getting the key fails.
func (e *E2e) GetHistoricalDHPubkey(js.Value, []js.Value) any {
	pubKey, err := e.api.GetHistoricalDHPubkey()
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}
	return utils.CopyBytesToJS(pubKey)
}

// HasAuthenticatedChannel returns true if an authenticated channel with the
// partner exists, otherwise returns false.
//
// Parameters:
//   - args[0] - Marshalled bytes of [id.ID] (Uint8Array).
//
// Returns:
//   - Existence of authenticated channel (boolean).
//   - Throws TypeError if unmarshalling the ID or getting the channel fails.
func (e *E2e) HasAuthenticatedChannel(_ js.Value, args []js.Value) any {
	exists, err := e.api.HasAuthenticatedChannel(utils.CopyBytesToGo(args[0]))
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}
	return exists
}

// RemoveService removes all services for the given tag.
//
// Parameters:
//   - args[0] - Tag of services to remove (string).
//
// Returns:
//   - Throws TypeError if removing the services fails.
func (e *E2e) RemoveService(_ js.Value, args []js.Value) any {
	err := e.api.RemoveService(args[0].String())
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return nil
}

// SendE2E send a message containing the payload to the recipient of the passed
// message type, per the given parameters--encrypted with end-to-end encryption.
//
// Parameters:
//   - args[0] - Message type from [catalog.MessageType] (int).
//   - args[1] - Marshalled bytes of [id.ID] (Uint8Array).
//   - args[2] - Message payload (Uint8Array).
//   - args[3] - JSON [xxdk.E2EParams] (Uint8Array).
//
// Returns a promise:
//   - Resolves to the JSON of the [bindings.E2ESendReport], which can be passed
//     into [Cmix.WaitForRoundResult] to see if the send succeeded (Uint8Array).
//   - Rejected with an error if sending fails.
func (e *E2e) SendE2E(_ js.Value, args []js.Value) any {
	mt := args[0].Int()
	recipientId := utils.CopyBytesToGo(args[1])
	payload := utils.CopyBytesToGo(args[2])
	e2eParams := utils.CopyBytesToGo(args[3])

	promiseFn := func(resolve, reject func(args ...any) js.Value) {
		sendReport, err := e.api.SendE2E(mt, recipientId, payload, e2eParams)
		if err != nil {
			reject(utils.JsTrace(err))
		} else {
			resolve(utils.CopyBytesToJS(sendReport))
		}
	}

	return utils.CreatePromise(promiseFn)
}

// processor wraps Javascript callbacks to adhere to the [bindings.Processor]
// interface.
type processor struct {
	process func(args ...any) js.Value
	string  func(args ...any) js.Value
}

// Process decrypts and hands off the message to its internal down stream
// message processing system.
//
// Parameters:
//   - message - Returns the message contents (Uint8Array).
//   - receptionId - Returns the marshalled bytes of the sender's [id.ID]
//     (Uint8Array).
//   - ephemeralId - Returns the ephemeral ID of the sender (int).
//   - roundId - Returns the ID of the round sent on (int).
func (p *processor) Process(
	message, receptionId []byte, ephemeralId, roundId int64) {
	p.process(utils.CopyBytesToJS(message), utils.CopyBytesToJS(receptionId),
		ephemeralId, roundId)
}

// String identifies this processor and is used for debugging.
func (p *processor) String() string {
	return p.string().String()
}

// AddService adds a service for all partners of the given tag, which will call
// back on the given processor. These can be sent to using the tag fields in the
// Params object.
//
// Passing nil for the processor allows you to create a service that is never
// called but will be visible by notifications. Processes added this way are
// generally not end-to-end encrypted messages themselves, but other protocols
// that piggyback on e2e relationships to start communication.
//
// Parameters:
//   - args[0] - tag for the service (string).
//   - args[1] - Javascript object that has functions that implement the
//     [bindings.Processor] interface.
//
// Returns:
//   - Throws TypeError if registering the service fails.
func (e *E2e) AddService(_ js.Value, args []js.Value) any {
	p := &processor{
		utils.WrapCB(args[1], "Process"), utils.WrapCB(args[1], "String")}

	err := e.api.AddService(args[0].String(), p)
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return nil
}

// RegisterListener registers a new listener.
//
// Parameters:
//   - args[0] - Marshalled byte of the user [id.ID] who sends messages to this
//     user that this function will register a listener for (Uint8Array).
//   - args[1] - Message type from [catalog.MessageType] you want to listen for
//     (int).
//   - args[2] - Javascript object that has functions that implement the
//     [bindings.Listener] interface; do not pass nil as the listener.
//
// Returns:
//   - Throws TypeError if registering the service fails.
func (e *E2e) RegisterListener(_ js.Value, args []js.Value) any {
	recipientId := utils.CopyBytesToGo(args[0])
	l := &listener{utils.WrapCB(args[2], "Hear"), utils.WrapCB(args[2], "Name")}

	err := e.api.RegisterListener(recipientId, args[1].Int(), l)
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return nil
}
