////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package worker

import (
	"github.com/hack-pad/safejs"
	"github.com/pkg/errors"

	"gitlab.com/elixxir/wasm-utils/utils"
)

// MessageChannel wraps a Javascript MessageChannel object.
//
// Doc: https://developer.mozilla.org/en-US/docs/Web/API/MessageChannel
type MessageChannel struct {
	safejs.Value
}

// NewMessageChannel returns a new MessageChannel object with two new
// MessagePort objects.
//
// Doc: https://developer.mozilla.org/en-US/docs/Web/API/MessageChannel/MessageChannel
func NewMessageChannel() (MessageChannel, error) {
	v, err := jsMessageChannel.New()
	if err != nil {
		return MessageChannel{}, err
	}
	return MessageChannel{v}, nil
}

// CreateMessageChannel creates a new Javascript MessageChannel between two
// workers. The [Channel] tag will be used as the prefix in the name of the
// MessageChannel when printing to logs. The key is used to look up the callback
// registered on the worker to handle the MessageChannel creation.
//
// Doc: https://developer.mozilla.org/en-US/docs/Web/API/MessageChannel
func CreateMessageChannel(w1, w2 *Manager, channelName, key string) error {
	// Create a Javascript MessageChannel
	mc, err := NewMessageChannel()
	if err != nil {
		return err
	}
	channelNameJS := utils.CopyBytesToJS([]byte(channelName))
	keyJS := utils.CopyBytesToJS([]byte(key))

	port1, err := mc.Port1()
	if err != nil {
		return errors.Wrap(err, "could not get port1")
	}

	port2, err := mc.Port2()
	if err != nil {
		return errors.Wrap(err, "could not get port2")
	}

	obj1 := map[string]any{
		"port": port1.Value, "channel": channelNameJS, "key": keyJS}
	err = w1.w.PostMessageTransfer(obj1, port1.Value)
	if err != nil {
		return errors.Wrap(err, "failed to send port1")
	}

	obj2 := map[string]any{
		"port": port2.Value, "channel": channelNameJS, "key": keyJS}
	err = w2.w.PostMessageTransfer(obj2, port2.Value)
	if err != nil {
		return errors.Wrap(err, "failed to send port2")
	}

	return nil
}

// Port1 returns the first port of the message channel — the port attached to
// the context that originated the channel.
//
// Doc: https://developer.mozilla.org/en-US/docs/Web/API/MessageChannel/port1
func (mc MessageChannel) Port1() (MessagePort, error) {
	v, err := mc.Get("port1")
	if err != nil {
		return MessagePort{}, err
	}
	return NewMessagePort(v)
}

// Port2 returns the second port of the message channel — the port attached to
// the context at the other end of the channel, which the message is initially
// sent to.
//
// Doc: https://developer.mozilla.org/en-US/docs/Web/API/MessageChannel/port2
func (mc MessageChannel) Port2() (MessagePort, error) {
	v, err := mc.Get("port2")
	if err != nil {
		return MessagePort{}, err
	}
	return NewMessagePort(v)
}
