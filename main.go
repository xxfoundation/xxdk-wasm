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
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"gitlab.com/elixxir/client/bindings"
	"gitlab.com/elixxir/xxdk-wasm/wasm"
	"os"
	"syscall/js"
)

func main() {
	fmt.Println("Starting xxDK WebAssembly bindings.")
	fmt.Printf("Client version %s\n", bindings.GetVersion())

	// wasm/backup.go
	js.Global().Set("NewCmixFromBackup", js.FuncOf(wasm.NewCmixFromBackup))
	js.Global().Set("InitializeBackup", js.FuncOf(wasm.InitializeBackup))
	js.Global().Set("ResumeBackup", js.FuncOf(wasm.ResumeBackup))

	// wasm/broadcast.go
	js.Global().Set("NewBroadcastChannel", js.FuncOf(wasm.NewBroadcastChannel))

	// wasm/cmix.go
	js.Global().Set("NewCmix", js.FuncOf(wasm.NewCmix))
	js.Global().Set("LoadCmix", js.FuncOf(wasm.LoadCmix))
	js.Global().Set("GetLoadCmix", js.FuncOf(wasm.GetLoadCmix))

	// wasm/dummy.go
	js.Global().Set("NewDummyTrafficManager",
		js.FuncOf(wasm.NewDummyTrafficManager))

	// wasm/e2e.go
	js.Global().Set("Login", js.FuncOf(wasm.Login))
	js.Global().Set("LoginEphemeral", js.FuncOf(wasm.LoginEphemeral))

	// wasm/errors.go
	js.Global().Set("CreateUserFriendlyErrorMessage",
		js.FuncOf(wasm.CreateUserFriendlyErrorMessage))
	js.Global().Set("UpdateCommonErrors",
		js.FuncOf(wasm.UpdateCommonErrors))

	// wasm/fileTransfer.go
	js.Global().Set("InitFileTransfer", js.FuncOf(wasm.InitFileTransfer))

	// wasm/group.go
	js.Global().Set("NewGroupChat", js.FuncOf(wasm.NewGroupChat))

	// wasm/identity.go
	js.Global().Set("StoreReceptionIdentity",
		js.FuncOf(wasm.StoreReceptionIdentity))
	js.Global().Set("LoadReceptionIdentity",
		js.FuncOf(wasm.LoadReceptionIdentity))
	js.Global().Set("GetContactFromReceptionIdentity",
		js.FuncOf(wasm.GetContactFromReceptionIdentity))
	js.Global().Set("GetIDFromContact",
		js.FuncOf(wasm.GetIDFromContact))
	js.Global().Set("GetPubkeyFromContact",
		js.FuncOf(wasm.GetPubkeyFromContact))
	js.Global().Set("SetFactsOnContact",
		js.FuncOf(wasm.SetFactsOnContact))
	js.Global().Set("GetFactsFromContact",
		js.FuncOf(wasm.GetFactsFromContact))

	// wasm/logging.go
	js.Global().Set("LogLevel", js.FuncOf(wasm.LogLevel))
	js.Global().Set("RegisterLogWriter", js.FuncOf(wasm.RegisterLogWriter))
	js.Global().Set("EnableGrpcLogs", js.FuncOf(wasm.EnableGrpcLogs))

	// wasm/ndf.go
	js.Global().Set("DownloadAndVerifySignedNdfWithUrl",
		js.FuncOf(wasm.DownloadAndVerifySignedNdfWithUrl))

	// wasm/params.go
	js.Global().Set("GetDefaultCMixParams",
		js.FuncOf(wasm.GetDefaultCMixParams))
	js.Global().Set("GetDefaultE2EParams",
		js.FuncOf(wasm.GetDefaultE2EParams))
	js.Global().Set("GetDefaultFileTransferParams",
		js.FuncOf(wasm.GetDefaultFileTransferParams))
	js.Global().Set("GetDefaultSingleUseParams",
		js.FuncOf(wasm.GetDefaultSingleUseParams))
	js.Global().Set("GetDefaultE2eFileTransferParams",
		js.FuncOf(wasm.GetDefaultE2eFileTransferParams))

	// wasm/restlike.go
	js.Global().Set("RestlikeRequest", js.FuncOf(wasm.RestlikeRequest))
	js.Global().Set("RestlikeRequestAuth", js.FuncOf(wasm.RestlikeRequestAuth))

	// wasm/restlikeSingle.go
	js.Global().Set("RequestRestLike",
		js.FuncOf(wasm.RequestRestLike))
	js.Global().Set("AsyncRequestRestLike",
		js.FuncOf(wasm.AsyncRequestRestLike))

	// wasm/secrets.go
	js.Global().Set("GenerateSecret", js.FuncOf(wasm.GenerateSecret))

	// wasm/single.go
	js.Global().Set("TransmitSingleUse", js.FuncOf(wasm.TransmitSingleUse))
	js.Global().Set("Listen", js.FuncOf(wasm.Listen))

	// wasm/ud.go
	js.Global().Set("NewOrLoadUd", js.FuncOf(wasm.NewOrLoadUd))
	js.Global().Set("NewUdManagerFromBackup",
		js.FuncOf(wasm.NewUdManagerFromBackup))
	js.Global().Set("LookupUD", js.FuncOf(wasm.LookupUD))
	js.Global().Set("SearchUD", js.FuncOf(wasm.SearchUD))

	// wasm/version.go
	js.Global().Set("GetVersion", js.FuncOf(wasm.GetVersion))
	js.Global().Set("GetGitVersion", js.FuncOf(wasm.GetGitVersion))
	js.Global().Set("GetDependencies", js.FuncOf(wasm.GetDependencies))

	defer func() {
		jww.CRITICAL.Printf("Before recover\n")
		if rec := recover(); rec != nil {
			wasm.Throw(wasm.TypeError, errors.Errorf(fmt.Sprintf("%+v", rec)))
		}
		jww.CRITICAL.Printf("After recover\n")
	}()

	<-make(chan bool)
	os.Exit(0)
}
