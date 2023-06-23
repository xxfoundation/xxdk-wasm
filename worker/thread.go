////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package worker

import (
	"syscall/js"
	"time"

	"github.com/hack-pad/safejs"
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
)

// ThreadReceptionCallback is called with a message received from the main
// thread. Any bytes returned are sent as a response back to the main thread.
// Any returned errors are printed to the log.
type ThreadReceptionCallback func(message []byte, reply func(message []byte))

// ThreadManager queues incoming messages from the main thread and handles them
// based on their tag.
type ThreadManager struct {
	mm *MessageManager

	// Wrapper of the DedicatedWorkerGlobalScope.
	// Doc: https://developer.mozilla.org/en-US/docs/Web/API/DedicatedWorkerGlobalScope
	t Thread
}

// NewThreadManager initialises a new ThreadManager.
func NewThreadManager(name string, messageLogging bool) (*ThreadManager, error) {
	t, err := NewThread()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to construct GlobalSelf")
	}

	p := DefaultParams()
	p.MessageLogging = messageLogging
	mm, err := NewMessageManager(t.Value, name+"-remote", p)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to construct message manager")
	}

	tm := &ThreadManager{
		mm: mm,
		t:  t,
	}

	return tm, nil
}

// Stop closes the thread manager and stops the worker.
func (tm *ThreadManager) Stop() error {
	tm.mm.Stop()

	// Close the worker
	err := tm.t.Close()
	return errors.Wrapf(err, "failed to close worker %q", tm.mm.name)
}

func (tm *ThreadManager) GetWorker() js.Value {
	return safejs.Unsafe(tm.t.Value)
}

// SignalReady sends a signal to the main thread indicating that the worker is
// ready. Once the main thread receives this, it will initiate communication.
// Therefore, this should only be run once all listeners are ready.
func (tm *ThreadManager) SignalReady() {
	err := tm.mm.SendNoResponse(readyTag, nil)
	if err != nil {
		jww.FATAL.Panicf(
			"[WW] [%s] Failed to send ready signal: %+v", tm.Name(), err)
	}
}

// RegisterMessageChannelCallback registers a callback that will be called when
// a MessagePort with the given Channel is received.
func (tm *ThreadManager) RegisterMessageChannelCallback(
	tag string, fn NewPortCallback) {
	tm.mm.RegisterMessageChannelCallback(tag, fn)
}

// SendMessage sends a message to the main thread with the given tag and waits
// for a response. An error is returned on failure to send or on timeout.
func (tm *ThreadManager) SendMessage(
	tag Tag, data []byte) (response []byte, err error) {
	return tm.mm.Send(tag, data)
}

// SendTimeout sends a message to the main thread with the given tag and waits
// for a response. An error is returned on failure to send or on the specified
// timeout.
func (tm *ThreadManager) SendTimeout(
	tag Tag, data []byte, timeout time.Duration) (response []byte, err error) {
	return tm.mm.SendTimeout(tag, data, timeout)
}

// SendNoResponse sends a message to the main thread with the given tag. It
// returns immediately and does not wait for a response.
func (tm *ThreadManager) SendNoResponse(tag Tag, data []byte) error {
	return tm.mm.SendNoResponse(tag, data)
}

// RegisterCallback registers the callback for the given tag. Previous tags are
// overwritten. This function is thread safe.
func (tm *ThreadManager) RegisterCallback(tag Tag, receiverCB ReceiverCallback) {
	tm.mm.RegisterCallback(tag, receiverCB)
}

// Name returns the name of the web worker.
func (tm *ThreadManager) Name() string { return tm.mm.name }

////////////////////////////////////////////////////////////////////////////////
// Javascript Call Wrappers                                                   //
////////////////////////////////////////////////////////////////////////////////

// Thread has the methods of the Javascript DedicatedWorkerGlobalScope.
//
// Doc: https://developer.mozilla.org/en-US/docs/Web/API/DedicatedWorkerGlobalScope
type Thread struct {
	MessagePort
}

// NewThread creates a new Thread from Global.
func NewThread() (Thread, error) {
	self, err := safejs.Global().Get("self")
	if err != nil {
		return Thread{}, err
	}

	mp, err := NewMessagePort(self)
	if err != nil {
		return Thread{}, err
	}

	return Thread{mp}, nil
}

// Name returns the name that the Worker was (optionally) given when it was
// created.
//
// Doc: https://developer.mozilla.org/en-US/docs/Web/API/DedicatedWorkerGlobalScope/name
func (t *Thread) Name() string {
	return safejs.Unsafe(t.Value).Get("name").String()
}

// Close discards any tasks queued in the worker's event loop, effectively
// closing this particular scope.
//
// Doc: https://developer.mozilla.org/en-US/docs/Web/API/DedicatedWorkerGlobalScope/close
func (t *Thread) Close() error {
	_, err := t.Call("close")
	return err
}
