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
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/xxdk-wasm/indexedDb"
)

func main() {
	fmt.Println("[WW] Starting xxDK WebAssembly DM Database Worker.")
	jww.SetStdoutThreshold(jww.LevelDebug)
	jww.INFO.Print("[WW] Starting xxDK WebAssembly DM Database Worker.")

	m := &manager{mh: indexedDb.NewMessageHandler("DmIndexedDbWorker")}
	m.RegisterHandlers()
	RegisterDatabaseNameStore(m)
	m.mh.SignalReady()
	<-make(chan bool)
	fmt.Println("[WW] Closing xxDK WebAssembly Channels Database Worker.")
}
