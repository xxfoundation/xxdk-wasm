////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package main

import (
	"fmt"
	"gitlab.com/elixxir/xxdk-wasm/indexedDb"
)

func main() {
	fmt.Println("Starting xxDK WebAssembly DM Database Worker.")

	m := &manager{mh: indexedDb.NewMessageHandler("DmIndexedDbWorker")}
	m.RegisterHandlers()
	RegisterDatabaseNameStore(m)
	m.mh.SignalReady()
	<-make(chan bool)
}
