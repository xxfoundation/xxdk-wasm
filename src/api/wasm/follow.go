////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"github.com/teamortix/golang-wasm/wasm"
	"gitlab.com/elixxir/xxdk-wasm/src/api/storage"
	"syscall/js"
)

// StartNetworkFollower kicks off the tracking of the network. It starts long-
// running network threads and returns an object for checking state and
// stopping those threads.
//
// Call this when returning from sleep and close when going back to sleep.
//
// These threads may become a significant drain on battery when offline, ensure
// they are stopped if there is no internet access.
//
// Threads Started:
//   - Network Follower (/network/follow.go)
//     tracks the network events and hands them off to workers for handling.
//   - Historical Round Retrieval (/network/rounds/historical.go)
//     retrieves data about rounds that are too old to be stored by the client.
//   - Message Retrieval Worker Group (/network/rounds/retrieve.go)
//     requests all messages in a given round from the gateway of the last
//     nodes.
//   - Message Handling Worker Group (/network/message/handle.go)
//     decrypts and partitions messages when signals via the Switchboard.
//   - Health Tracker (/network/health),
//     via the network instance, tracks the state of the network.
//   - Garbled Messages (/network/message/garbled.go)
//     can be signaled to check all recent messages that could be decoded. It
//     uses a message store on disk for persistence.
//   - Critical Messages (/network/message/critical.go)
//     ensures all protocol layer mandatory messages are sent. It uses a message
//     store on disk for persistence.
//   - KeyExchange Trigger (/keyExchange/trigger.go)
//     responds to sent rekeys and executes them.
//   - KeyExchange Confirm (/keyExchange/confirm.go)
//     responds to confirmations of successful rekey operations.
//   - Auth Callback (/auth/callback.go)
//     handles both auth confirm and requests.
//
// Parameters:
//   - timeoutMS - Timeout when stopping threads in milliseconds (int).
func (c *Cmix) StartNetworkFollower(timeoutMS int) error {
	err := c.Cmix.StartNetworkFollower(timeoutMS)
	if err != nil {
		return err
	}

	storage.IncrementNumClientsRunning()
	return nil
}

// StopNetworkFollower stops the network follower if it is running. It returns
// an error if the follower is in the wrong state to stop or if it fails to stop
// it.
//
// If the network follower is running and this fails, the [Cmix] object will
// most likely be in an unrecoverable state and need to be trashed.
func (c *Cmix) StopNetworkFollower() error {
	err := c.Cmix.StopNetworkFollower()
	if err != nil {
		return err
	}

	storage.DecrementNumClientsRunning()
	return nil
}

// NetworkHealthCallback wraps a Javascript object so that it implements the
// [bindings.NetworkHealthCallback] interface.
type NetworkHealthCallback struct {
	CallbackFn func(health bool) `wasm:"Callback"`
}

// Callback receives notification if network health changes.
//
// Parameters:
//   - health - Returns true if the network is healthy and false otherwise.
func (nhc *NetworkHealthCallback) Callback(health bool) {
	nhc.CallbackFn(health)
}

// AddHealthCallback adds a callback that gets called whenever the network
// health changes. Returns a registration ID that can be used to unregister.
//
// Parameters:
//   - nhc - Javascript object that matches the [NetworkHealthCallback] struct.
//
// Returns:
//   - A registration ID that can be used to unregister the callback.
func (c *Cmix) AddHealthCallback(nhc *NetworkHealthCallback) int64 {
	return c.Cmix.AddHealthCallback(nhc)
}

// ClientError wraps a Javascript object so that it implements the
// [bindings.ClientError] interface.
type ClientError struct {
	ReportFn func(source, message, trace string) `wasm:"Report"`
}

// Report handles errors from the network follower threads.
func (ce *ClientError) Report(source, message, trace string) {
	ce.ReportFn(source, message, trace)
}

// RegisterClientErrorCallback registers the callback to handle errors from the
// long-running threads controlled by StartNetworkFollower and
// StopNetworkFollower.
//
// Parameters:
//   - clientError - Javascript object that matches the [ClientError] struct.
func (c *Cmix) RegisterClientErrorCallback(clientError *ClientError) {
	c.Cmix.RegisterClientErrorCallback(clientError)
}

// TrackServicesCallback wraps a Javascript object so that it implements the
// [bindings.TrackServicesCallback] interface.
type TrackServicesCallback struct {
	CallbackFn func(marshalData []byte, err js.Value) `wasm:"Callback"`
}

// Callback is the callback for [Cmix.TrackServices]. This will pass to the user
// a JSON-marshalled list of backend services. If there was an error retrieving
// or marshalling the service list, there is an error for the second parameter,
// which will be non-null.
//
// Parameters:
//   - marshalData - Returns the JSON of [message.ServiceList] (Uint8Array).
//   - err - Returns an error on failure (Error).
//
// Example JSON:
//
//	[
//	  {
//	    "Id": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAD", // bytes of id.ID encoded as base64 string
//	    "Services": [
//	      {
//	        "Identifier": "AQID",                             // bytes encoded as base64 string
//	        "Tag": "TestTag 1",                               // string
//	        "Metadata": "BAUG"                                // bytes encoded as base64 string
//	      }
//	    ]
//	  },
//	  {
//	    "Id": "AAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAD",
//	    "Services": [
//	      {
//	        "Identifier": "AQID",
//	        "Tag": "TestTag 2",
//	        "Metadata": "BAUG"
//	      }
//	    ]
//	  },
//	]
func (tsc *TrackServicesCallback) Callback(marshalData []byte, err error) {
	tsc.CallbackFn(marshalData, wasm.NewError(err))
}

// TrackServicesWithIdentity will return via a callback the list of services the
// backend keeps track of for the provided identity. This may be passed into
// other bindings call which may need context on the available services for this
// single identity. This will only return services for the given identity.
//
// Parameters:
//   - e2eID - ID of [E2e] object in tracker.
//   - cb - Javascript object that matches the [TrackServicesCallback] struct.
func (c *Cmix) TrackServicesWithIdentity(
	e2eID int, cb *TrackServicesCallback) error {
	return c.Cmix.TrackServicesWithIdentity(e2eID, cb)
}

// TrackServices will return, via a callback, the list of services that the
// backend keeps track of, which is formally referred to as a
// [message.ServiceList]. This may be passed into other bindings call that may
// need context on the available services for this client.
//
// Parameters:
//   - cb - Javascript object that matches the [TrackServicesCallback] struct.
func (c *Cmix) TrackServices(cb *TrackServicesCallback) {
	c.Cmix.TrackServices(cb)
}
