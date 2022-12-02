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
)

func main() {
	fmt.Println("Starting xxDK WebAssembly Database Worker.")

	mh := newMessageHandler()
	mh.registerHandlers()
	registerDatabaseNameStore(mh)
	mh.signalReady()
	<-make(chan bool)
}
