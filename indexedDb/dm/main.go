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
	"gitlab.com/elixxir/xxdk-wasm/wasm"
	"syscall/js"
)

func init() {
	// Set up Javascript console listener set at level INFO
	ll := wasm.NewJsConsoleLogListener(jww.LevelInfo)
	jww.SetLogListeners(ll.Listen)
	jww.SetStdoutThreshold(jww.LevelFatal + 1)
}

func main() {
	fmt.Println("[WW] Starting xxDK WebAssembly DM Database Worker.")
	jww.INFO.Print("[WW] Starting xxDK WebAssembly DM Database Worker.")

	js.Global().Set("LogLevel", js.FuncOf(wasm.LogLevel))
	js.Global().Set("LogToFile", js.FuncOf(wasm.LogToFile))
	js.Global().Set("RegisterLogWriter", js.FuncOf(wasm.RegisterLogWriter))

	m := &manager{mh: indexedDb.NewMessageHandler("DmIndexedDbWorker")}
	m.RegisterHandlers()
	m.mh.SignalReady()
	<-make(chan bool)
	fmt.Println("[WW] Closing xxDK WebAssembly Channels Database Worker.")
}
