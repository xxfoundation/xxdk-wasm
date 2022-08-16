///////////////////////////////////////////////////////////////////////////////
// Copyright © 2020 xx network SEZC                                          //
//                                                                           //
// Use of this source code is governed by a license that can be found in the //
// LICENSE file                                                              //
///////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package main

import (
	"fmt"
	"gitlab.com/elixxir/xxdk-wasm/bindings"
	"os"
	"os/signal"
	"syscall"
	"syscall/js"
)

func main() {
	fmt.Println("Go Web Assembly")

	// wasm/cmix.go
	js.Global().Set("NewCmix", js.FuncOf(bindings.NewCmix))
	js.Global().Set("LoadCmix", js.FuncOf(bindings.LoadCmix))

	// wasm/e2e.go
	js.Global().Set("Login", js.FuncOf(bindings.Login))
	js.Global().Set("LoginEphemeral", js.FuncOf(bindings.LoginEphemeral))

	// wasm/identity.go
	js.Global().Set("StoreReceptionIdentity",
		js.FuncOf(bindings.StoreReceptionIdentity))
	js.Global().Set("LoadReceptionIdentity",
		js.FuncOf(bindings.LoadReceptionIdentity))
	js.Global().Set("GetIDFromContact",
		js.FuncOf(bindings.GetIDFromContact))
	js.Global().Set("GetPubkeyFromContact",
		js.FuncOf(bindings.GetPubkeyFromContact))
	js.Global().Set("SetFactsOnContact",
		js.FuncOf(bindings.SetFactsOnContact))
	js.Global().Set("GetFactsFromContact",
		js.FuncOf(bindings.GetFactsFromContact))

	// wasm/params.go
	js.Global().Set("GetDefaultCMixParams",
		js.FuncOf(bindings.GetDefaultCMixParams))
	js.Global().Set("GetDefaultE2EParams",
		js.FuncOf(bindings.GetDefaultE2EParams))
	js.Global().Set("GetDefaultFileTransferParams",
		js.FuncOf(bindings.GetDefaultFileTransferParams))
	js.Global().Set("GetDefaultSingleUseParams",
		js.FuncOf(bindings.GetDefaultSingleUseParams))
	js.Global().Set("GetDefaultE2eFileTransferParams",
		js.FuncOf(bindings.GetDefaultE2eFileTransferParams))

	// wasm/logging.go
	js.Global().Set("LogLevel", js.FuncOf(bindings.LogLevel))
	js.Global().Set("RegisterLogWriter", js.FuncOf(bindings.RegisterLogWriter))
	js.Global().Set("EnableGrpcLogs", js.FuncOf(bindings.EnableGrpcLogs))

	// wasm/ndf.go
	js.Global().Set("DownloadAndVerifySignedNdfWithUrl",
		js.FuncOf(bindings.DownloadAndVerifySignedNdfWithUrl))

	// wasm/version.go
	js.Global().Set("GetVersion", js.FuncOf(bindings.GetVersion))
	js.Global().Set("GetGitVersion", js.FuncOf(bindings.GetGitVersion))
	js.Global().Set("GetDependencies", js.FuncOf(bindings.GetDependencies))

	// wasm/secrets.go
	js.Global().Set("GenerateSecret", js.FuncOf(bindings.GenerateSecret))

	// wasm/dummy.go
	js.Global().Set("NewDummyTrafficManager",
		js.FuncOf(bindings.NewDummyTrafficManager))

	// Wait until the user terminates the program
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	os.Exit(0)
}
