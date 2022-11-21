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
	"gitlab.com/elixxir/client/bindings"
	"gitlab.com/elixxir/xxdk-wasm/storage"
	"gitlab.com/elixxir/xxdk-wasm/utils"
	"gitlab.com/elixxir/xxdk-wasm/wasm"
	"os"
	"syscall/js"
)

func init() {
	// Overwrites setting the log level to INFO done in bindings so that the
	// Javascript console can be used
	ll := wasm.NewJsConsoleLogListener(jww.LevelInfo)
	jww.SetLogListeners(ll.Listen)
	jww.SetStdoutThreshold(jww.LevelFatal + 1)

	// Check that the WASM binary version is correct
	err := storage.CheckAndStoreVersions()
	if err != nil {
		jww.FATAL.Panicf("WASM binary version error: %+v", err)
	}
}

func main() {
	fmt.Println("Starting xxDK WebAssembly bindings.")
	fmt.Printf("Client version %s\n", bindings.GetVersion())

	// storage/password.go
	js.Global().Set("GetOrInitPassword", js.FuncOf(storage.GetOrInitPassword))
	js.Global().Set("ChangeExternalPassword",
		js.FuncOf(storage.ChangeExternalPassword))
	js.Global().Set("VerifyPassword", js.FuncOf(storage.VerifyPassword))

	// storage/purge.go
	js.Global().Set("Purge", js.FuncOf(storage.Purge))

	// utils/array.go
	js.Global().Set("Uint8ArrayToBase64", js.FuncOf(utils.Uint8ArrayToBase64))
	js.Global().Set("Base64ToUint8Array", js.FuncOf(utils.Base64ToUint8Array))
	js.Global().Set("Uint8ArrayEquals", js.FuncOf(utils.Uint8ArrayEquals))

	// wasm/backup.go
	js.Global().Set("NewCmixFromBackup", js.FuncOf(wasm.NewCmixFromBackup))
	js.Global().Set("InitializeBackup", js.FuncOf(wasm.InitializeBackup))
	js.Global().Set("ResumeBackup", js.FuncOf(wasm.ResumeBackup))

	// wasm/channels.go
	js.Global().Set("GenerateChannelIdentity",
		js.FuncOf(wasm.GenerateChannelIdentity))
	js.Global().Set("ConstructIdentity", js.FuncOf(wasm.ConstructIdentity))
	js.Global().Set("ImportPrivateIdentity",
		js.FuncOf(wasm.ImportPrivateIdentity))
	js.Global().Set("GetPublicChannelIdentity",
		js.FuncOf(wasm.GetPublicChannelIdentity))
	js.Global().Set("GetPublicChannelIdentityFromPrivate",
		js.FuncOf(wasm.GetPublicChannelIdentityFromPrivate))
	js.Global().Set("NewChannelsManager", js.FuncOf(wasm.NewChannelsManager))
	js.Global().Set("LoadChannelsManager", js.FuncOf(wasm.LoadChannelsManager))
	js.Global().Set("NewChannelsManagerWithIndexedDb",
		js.FuncOf(wasm.NewChannelsManagerWithIndexedDb))
	js.Global().Set("LoadChannelsManagerWithIndexedDb",
		js.FuncOf(wasm.LoadChannelsManagerWithIndexedDb))
	js.Global().Set("LoadChannelsManagerWithIndexedDbUnsafe",
		js.FuncOf(wasm.LoadChannelsManagerWithIndexedDbUnsafe))
	js.Global().Set("NewChannelsManagerWithIndexedDbUnsafe",
		js.FuncOf(wasm.NewChannelsManagerWithIndexedDbUnsafe))
	js.Global().Set("GenerateChannel", js.FuncOf(wasm.GenerateChannel))
	js.Global().Set("GetSavedChannelPrivateKey",
		js.FuncOf(wasm.GetSavedChannelPrivateKey))
	js.Global().Set("ImportChannelPrivateKey",
		js.FuncOf(wasm.ImportChannelPrivateKey))
	js.Global().Set("GetSavedChannelPrivateKeyUNSAFE",
		js.FuncOf(wasm.GetSavedChannelPrivateKeyUNSAFE))
	js.Global().Set("DecodePublicURL", js.FuncOf(wasm.DecodePublicURL))
	js.Global().Set("DecodePrivateURL", js.FuncOf(wasm.DecodePrivateURL))
	js.Global().Set("GetChannelJSON", js.FuncOf(wasm.GetChannelJSON))
	js.Global().Set("GetChannelInfo", js.FuncOf(wasm.GetChannelInfo))
	js.Global().Set("GetShareUrlType", js.FuncOf(wasm.GetShareUrlType))
	js.Global().Set("IsNicknameValid", js.FuncOf(wasm.IsNicknameValid))
	js.Global().Set("NewChannelsDatabaseCipher",
		js.FuncOf(wasm.NewChannelsDatabaseCipher))

	// wasm/cmix.go
	js.Global().Set("NewCmix", js.FuncOf(wasm.NewCmix))
	js.Global().Set("LoadCmix", js.FuncOf(wasm.LoadCmix))

	// wasm/delivery.go
	js.Global().Set("SetDashboardURL", js.FuncOf(wasm.SetDashboardURL))

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
	js.Global().Set("DeserializeGroup", js.FuncOf(wasm.DeserializeGroup))

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
	js.Global().Set("LogToFile", js.FuncOf(wasm.LogToFile))
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

	// wasm/timeNow.go
	js.Global().Set("SetTimeSource", js.FuncOf(wasm.SetTimeSource))
	js.Global().Set("SetOffset", js.FuncOf(wasm.SetOffset))

	// wasm/ud.go
	js.Global().Set("NewOrLoadUd", js.FuncOf(wasm.NewOrLoadUd))
	js.Global().Set("NewUdManagerFromBackup",
		js.FuncOf(wasm.NewUdManagerFromBackup))
	js.Global().Set("LookupUD", js.FuncOf(wasm.LookupUD))
	js.Global().Set("SearchUD", js.FuncOf(wasm.SearchUD))

	// wasm/version.go
	js.Global().Set("GetVersion", js.FuncOf(wasm.GetVersion))
	js.Global().Set("GetClientVersion", js.FuncOf(wasm.GetClientVersion))
	js.Global().Set("GetClientGitVersion", js.FuncOf(wasm.GetClientGitVersion))
	js.Global().Set("GetClientDependencies", js.FuncOf(wasm.GetClientDependencies))

	<-make(chan bool)
	os.Exit(0)
}
