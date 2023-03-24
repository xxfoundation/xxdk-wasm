////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package logging

import (
	jww "github.com/spf13/jwalterweatherman"
	"sync"
)

// logListeners contains a map of all registered log listeners keyed on a unique
// ID that can be used to remove the listener once it has been added. This
// global keeps track of all listeners that are registered to jwalterweatherman
// logging.
var logListeners = newLogListenerList()

type logListenerList struct {
	listeners map[uint64]jww.LogListener
	currentID uint64
	sync.Mutex
}

func newLogListenerList() logListenerList {
	return logListenerList{
		listeners: make(map[uint64]jww.LogListener),
		currentID: 0,
	}
}

// AddLogListener registers the log listener with jwalterweatherman. Returns a
// unique ID that can be used to remove the listener.
func AddLogListener(ll jww.LogListener) uint64 {
	logListeners.Lock()
	defer logListeners.Unlock()

	id := logListeners.addLogListener(ll)
	jww.SetLogListeners(logListeners.mapToSlice()...)
	return id
}

// RemoveLogListener unregisters the log listener with the ID from
// jwalterweatherman.
func RemoveLogListener(id uint64) {
	logListeners.Lock()
	defer logListeners.Unlock()

	logListeners.removeLogListener(id)
	jww.SetLogListeners(logListeners.mapToSlice()...)

}

// addLogListener adds the listener to the list and returns its unique ID.
func (lll *logListenerList) addLogListener(ll jww.LogListener) uint64 {
	id := lll.currentID
	lll.currentID++
	lll.listeners[id] = ll

	return id
}

// removeLogListener removes the listener with the specified ID from the list.
func (lll *logListenerList) removeLogListener(id uint64) {
	delete(lll.listeners, id)
}

// mapToSlice converts the map of listeners to a slice of listeners so that it
// can be registered with jwalterweatherman.SetLogListeners.
func (lll *logListenerList) mapToSlice() []jww.LogListener {
	listeners := make([]jww.LogListener, 0, len(lll.listeners))
	for _, l := range lll.listeners {
		listeners = append(listeners, l)
	}
	return listeners
}
