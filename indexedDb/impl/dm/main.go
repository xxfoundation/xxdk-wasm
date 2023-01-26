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
	"gitlab.com/elixxir/xxdk-wasm/logging"
	"gitlab.com/elixxir/xxdk-wasm/wasm"
	"gitlab.com/elixxir/xxdk-wasm/worker"
	"syscall/js"
)

// SEMVER is the current semantic version of the xxDK DM web worker.
const SEMVER = "0.1.0"

func init() {
	// Set up Javascript console listener set at level INFO
	ll := logging.NewJsConsoleLogListener(jww.LevelInfo)
	logging.AddLogListener(ll.Listen)
	jww.SetStdoutThreshold(jww.LevelFatal + 1)
	jww.INFO.Printf("xxDK DM web worker version: v%s", SEMVER)
}

func main() {
	jww.INFO.Print("[WW] Starting xxDK WebAssembly DM Database Worker.")

	js.Global().Set("LogLevel", js.FuncOf(wasm.LogLevel))

	m := &manager{mh: worker.NewThreadManager("DmIndexedDbWorker", true)}
	m.registerCallbacks()
	m.mh.SignalReady()
	<-make(chan bool)
	fmt.Println("[WW] Closing xxDK WebAssembly Channels Database Worker.")
}
