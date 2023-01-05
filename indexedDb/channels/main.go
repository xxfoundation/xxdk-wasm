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
	fmt.Println("Starting xxDK WebAssembly Database Worker.")

	m := &manager{mh: indexedDb2.NewMessageHandler()}
	RegisterDatabaseNameStore(m)
	m.RegisterHandlers()
	m.mh.SignalReady()
	<-make(chan bool)
}
