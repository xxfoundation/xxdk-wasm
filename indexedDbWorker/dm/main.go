////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package main

import (
	"fmt"
	"gitlab.com/elixxir/xxdk-wasm/indexedDbWorker"
)

func main() {
	fmt.Println("Starting xxDK WebAssembly Database Worker.")

	m := &manager{mh: indexedDbWorker.NewMessageHandler()}
	m.RegisterHandlers()
	RegisterDatabaseNameStore(m)
	m.mh.SignalReady()
	<-make(chan bool)
}
