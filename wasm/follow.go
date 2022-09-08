////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package wasm

import (
	"gitlab.com/elixxir/xxdk-wasm/utils"
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
//	 - Message Retrieval Worker Group (/network/rounds/retrieve.go)
//	   requests all messages in a given round from the gateway of the last
//	   nodes.
//	 - Message Handling Worker Group (/network/message/handle.go)
//	   decrypts and partitions messages when signals via the Switchboard.
//	 - Health Tracker (/network/health),
//	   via the network instance, tracks the state of the network.
//	 - Garbled Messages (/network/message/garbled.go)
//	   can be signaled to check all recent messages that could be decoded. It
//	   uses a message store on disk for persistence.
//	 - Critical Messages (/network/message/critical.go)
//	   ensures all protocol layer mandatory messages are sent. It uses a message
//	   store on disk for persistence.
//	 - KeyExchange Trigger (/keyExchange/trigger.go)
//	   responds to sent rekeys and executes them.
//   - KeyExchange Confirm (/keyExchange/confirm.go)
//	   responds to confirmations of successful rekey operations.
//   - Auth Callback (/auth/callback.go)
//     handles both auth confirm and requests.
//
// Parameters:
//  - args[0] - timeout when stopping threads in milliseconds (int)
//
// Returns:
//  - throws a TypeError if starting the network follower fails
func (c *Cmix) StartNetworkFollower(_ js.Value, args []js.Value) interface{} {
	err := c.api.StartNetworkFollower(args[0].Int())
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return nil
}

// StopNetworkFollower stops the network follower if it is running.
//
// If the network follower is running and this fails, the Cmix object will
// most likely be in an unrecoverable state and need to be trashed.
//
// Returns:
//  - throws a TypeError if the follower is in the wrong state to stop or if it
//    fails to stop
func (c *Cmix) StopNetworkFollower(js.Value, []js.Value) interface{} {
	err := c.api.StopNetworkFollower()
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return nil
}

// WaitForNetwork will block until either the network is healthy or the passed
// timeout is reached. It will return true if the network is healthy.
//
// Parameters:
//  - args[0] - timeout when stopping threads in milliseconds (int)
//
// Returns:
//  - returns true if the network is healthy (boolean)
func (c *Cmix) WaitForNetwork(_ js.Value, args []js.Value) interface{} {
	return c.api.WaitForNetwork(args[0].Int())
}

// NetworkFollowerStatus gets the state of the network follower. It returns a
// status with the following values:
//  Stopped  - 0
//  Running  - 2000
//  Stopping - 3000
//
// Returns:
//  - returns network status code (int)
func (c *Cmix) NetworkFollowerStatus(js.Value, []js.Value) interface{} {
	return c.api.NetworkFollowerStatus()
}

// GetNodeRegistrationStatus returns the current state of node registration.
//
// Returns:
//  - []byte - JSON of [bindings.NodeRegistrationReport] containing the number
//    of nodes the user is registered with and the number of nodes present in
//    the NDF.
//  - An error if it cannot get the node registration status. The most likely
//    cause is that the network is unhealthy.
func (c *Cmix) GetNodeRegistrationStatus(js.Value, []js.Value) interface{} {
	b, err := c.api.GetNodeRegistrationStatus()
	if err != nil {
		utils.Throw(utils.TypeError, err)
		return nil
	}

	return utils.CopyBytesToJS(b)
}

// HasRunningProcessies checks if any background threads are running and returns
// true if one or more are.
//
// This is meant to be used when NetworkFollowerStatus returns Stopping. Due to
// the handling of comms on iOS, where the OS can block indefinitely, it may not
// enter the stopped state appropriately. This can be used instead.
//
// Returns:
//  - boolean
func (c *Cmix) HasRunningProcessies(js.Value, []js.Value) interface{} {
	return c.api.HasRunningProcessies()
}

// IsHealthy returns true if the network is read to be in a healthy state where
// messages can be sent.
//
// Returns:
//  - boolean
func (c *Cmix) IsHealthy(js.Value, []js.Value) interface{} {
	return c.api.IsHealthy()
}

// networkHealthCallback adheres to the [bindings.NetworkHealthCallback]
// interface.
type networkHealthCallback struct {
	callback func(args ...interface{}) js.Value
}

func (nhc *networkHealthCallback) Callback(health bool) { nhc.callback(health) }

// AddHealthCallback adds a callback that gets called whenever the network
// health changes. Returns a registration ID that can be used to unregister.
//
// Parameters:
//  - args[0] - Javascript object that has functions that implement the
//    [bindings.NetworkHealthCallback] interface
//
// Returns:
//  - Returns a registration ID that can be used to unregister (int)
func (c *Cmix) AddHealthCallback(_ js.Value, args []js.Value) interface{} {
	return c.api.AddHealthCallback(
		&networkHealthCallback{utils.WrapCB(args[0], "Callback")})
}

// RemoveHealthCallback removes a health callback using its registration ID.
//
// Parameters:
//  - args[0] - registration ID (int)
func (c *Cmix) RemoveHealthCallback(_ js.Value, args []js.Value) interface{} {
	c.api.RemoveHealthCallback(int64(args[0].Int()))
	return nil
}

// clientError adheres to the [bindings.ClientError] interface.
type clientError struct {
	report func(args ...interface{}) js.Value
}

func (ce *clientError) Report(source, message, trace string) {
	ce.report(source, message, trace)
}

// RegisterClientErrorCallback registers the callback to handle errors from the
// long-running threads controlled by StartNetworkFollower and
// StopNetworkFollower.
//
// Parameters:
//  - args[0] - Javascript object that has functions that implement the
//    [bindings.ClientError] interface
func (c *Cmix) RegisterClientErrorCallback(_ js.Value, args []js.Value) interface{} {
	c.api.RegisterClientErrorCallback(&clientError{utils.WrapCB(args[0], "Report")})
	return nil
}
