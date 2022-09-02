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
)

// GetReceptionID returns the marshalled default IDs.
//
// Returns:
//  - The marshalled bytes of the [id.ID] object (Uint8Array)
func (e *E2e) GetReceptionID(js.Value, []js.Value) interface{} {
	return CopyBytesToJS(e.api.GetReceptionID())
}

// GetAllPartnerIDs returns a list of all partner IDs that the user has an E2E
// relationship with.
//
// Returns:
//  - JSON of array of [id.ID] (Uint8Array)
//  - Throws TypeError if getting partner IDs fails
func (e *E2e) GetAllPartnerIDs(js.Value, []js.Value) interface{} {
	partnerIDs, err := e.api.GetAllPartnerIDs()
	if err != nil {
		Throw(TypeError, err)
		return nil
	}
	return CopyBytesToJS(partnerIDs)
}

// PayloadSize returns the max payload size for a partitionable E2E message.
//
// Returns:
//  - Max payload size (int)
func (e *E2e) PayloadSize(js.Value, []js.Value) interface{} {
	return e.api.PayloadSize()
}

// SecondPartitionSize returns the max partition payload size for all payloads
// after the first payload.
//
// Returns:
//  - Max payload size (int)
func (e *E2e) SecondPartitionSize(js.Value, []js.Value) interface{} {
	return e.api.SecondPartitionSize()
}

// PartitionSize returns the partition payload size for the given payload index.
// The first payload is index 0.
//
// Parameters:
//  - args[0] - payload index (int)
//
// Returns:
//  - Partition payload size (int)
func (e *E2e) PartitionSize(_ js.Value, args []js.Value) interface{} {
	return e.api.PartitionSize(args[0].Int())
}

// FirstPartitionSize returns the max partition payload size for the first
// payload.
//
// Returns:
//  - Max partition payload size (int)
func (e *E2e) FirstPartitionSize(js.Value, []js.Value) interface{} {
	return e.api.FirstPartitionSize()
}

// GetHistoricalDHPrivkey returns the user's marshalled historical DH private
// key.
//
// Returns:
//  - JSON of [cyclic.Int] (Uint8Array)
//  - Throws TypeError if getting the key fails
func (e *E2e) GetHistoricalDHPrivkey(js.Value, []js.Value) interface{} {
	privKey, err := e.api.GetHistoricalDHPrivkey()
	if err != nil {
		Throw(TypeError, err)
		return nil
	}
	return CopyBytesToJS(privKey)
}

// GetHistoricalDHPubkey returns the user's marshalled historical DH public key.
// key.
//
// Returns:
//  - JSON of [cyclic.Int] (Uint8Array)
//  - Throws TypeError if getting the key fails
func (e *E2e) GetHistoricalDHPubkey(js.Value, []js.Value) interface{} {
	pubKey, err := e.api.GetHistoricalDHPubkey()
	if err != nil {
		Throw(TypeError, err)
		return nil
	}
	return CopyBytesToJS(pubKey)
}

// HasAuthenticatedChannel returns true if an authenticated channel with the
// partner exists, otherwise returns false.
//
// Parameters:
//  - args[0] - JSON of [id.ID] (Uint8Array)
//
// Returns:
//  - Existence of authenticated channel (boolean)
//  - Throws TypeError if unmarshalling the ID or getting the channel fails
func (e *E2e) HasAuthenticatedChannel(_ js.Value, args []js.Value) interface{} {
	exists, err := e.api.HasAuthenticatedChannel(CopyBytesToGo(args[0]))
	if err != nil {
		Throw(TypeError, err)
		return nil
	}
	return exists
}

// RemoveService removes all services for the given tag.
//
// Parameters:
//  - args[0] - tag of services to remove (string)
//
// Returns:
//  - Throws TypeError if removing the services fails
func (e *E2e) RemoveService(_ js.Value, args []js.Value) interface{} {
	err := e.api.RemoveService(args[0].String())
	if err != nil {
		Throw(TypeError, err)
		return nil
	}

	return nil
}

// SendE2E send a message containing the payload to the recipient of the passed
// message type, per the given parameters--encrypted with end-to-end encryption.
//
// Parameters:
//  - args[0] - message type from [catalog.MessageType] (int)
//  - args[1] - JSON of [id.ID] (Uint8Array)
//  - args[2] - message payload (Uint8Array)
//  - args[3] - JSON [e2e.Params] (Uint8Array)
//
// Returns:
//  - JSON of the [bindings.E2ESendReport], which can be passed into
//    Cmix.WaitForRoundResult to see if the send succeeded (Uint8Array)
//  - Throws TypeError if sending fails
func (e *E2e) SendE2E(_ js.Value, args []js.Value) interface{} {
	recipientId := CopyBytesToGo(args[1])
	payload := CopyBytesToGo(args[2])
	e2eParams := CopyBytesToGo(args[3])

	sendReport, err := e.api.SendE2E(
		args[0].Int(), recipientId, payload, e2eParams)
	if err != nil {
		Throw(TypeError, err)
		return nil
	}

	return CopyBytesToJS(sendReport)
}

// processor wraps Javascript callbacks to adhere to the [bindings.Processor]
// interface.
type processor struct {
	process func(args ...interface{}) js.Value
	string  func(args ...interface{}) js.Value
}

func (p *processor) Process(
	message, receptionId []byte, ephemeralId, roundId int64) {
	p.process(CopyBytesToJS(message), CopyBytesToJS(receptionId), ephemeralId,
		roundId)
}

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
//  - args[0] - tag for the service (string)
//  - args[1] - Javascript object that has functions that implement the
//    [bindings.Processor] interface
//
// Returns:
//  - Throws TypeError if registering the service fails
func (e *E2e) AddService(_ js.Value, args []js.Value) interface{} {
	p := &processor{WrapCB(args[1], "Process"), WrapCB(args[1], "String")}

	err := e.api.AddService(args[0].String(), p)
	if err != nil {
		Throw(TypeError, err)
		return nil
	}

	return nil
}

// RegisterListener registers a new listener.
//
// Parameters:
//  - args[0] - JSON of the user ID [id.ID] who sends messages to this user that
//    this function will register a listener for (Uint8Array)
//  - args[1] - message type from [catalog.MessageType] you want to listen for
//    (int)
//  - args[2] - Javascript object that has functions that implement the
//    [bindings.Listener] interface; do not pass nil as the listener
//
// Returns:
//  - Throws TypeError if registering the service fails
func (e *E2e) RegisterListener(_ js.Value, args []js.Value) interface{} {
	recipientId := CopyBytesToGo(args[0])
	l := &listener{WrapCB(args[1], "Hear"), WrapCB(args[1], "Name")}

	err := e.api.RegisterListener(recipientId, args[1].Int(), l)
	if err != nil {
		Throw(TypeError, err)
		return nil
	}

	return nil
}
