////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package main

import (
	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	wasm2 "github.com/teamortix/golang-wasm/wasm"
	"gitlab.com/elixxir/client/v4/bindings"
	"os"

	"gitlab.com/elixxir/xxdk-wasm/src/api/logging"
	"gitlab.com/elixxir/xxdk-wasm/src/api/storage"
	"gitlab.com/elixxir/xxdk-wasm/src/api/utils"
	"gitlab.com/elixxir/xxdk-wasm/src/api/wasm"
)

// Automatically handled with Promise rejects when returning an error!
func divide(x int, y int) (int, error) {
	if y == 0 {
		return 0, errors.New("cannot divide by zero")
	}
	return x / y, nil
}

func init() {
	// Start logger first to capture all logging events
	logging.InitLogger()

	// Overwrites setting the log level to INFO done in bindings so that the
	// Javascript console can be used
	ll := logging.NewJsConsoleLogListener(jww.LevelInfo)
	logging.AddLogListener(ll.Listen)
	jww.SetStdoutThreshold(jww.LevelFatal + 1)

	// Check that the WASM binary version is correct
	err := storage.CheckAndStoreVersions()
	if err != nil {
		jww.FATAL.Panicf("WASM binary version error: %+v", err)
	}
}

func main() {
	jww.INFO.Printf("Starting xxDK WebAssembly bindings.")

	wasm2.Expose("divide", divide)

	// logging/worker.go
	wasm2.Expose("GetLogger", logging.GetLoggerJS)

	// storage/password.go
	wasm2.Expose("GetOrInitPassword", storage.GetOrInitPassword)
	wasm2.Expose("ChangeExternalPassword",
		storage.ChangeExternalPassword)
	wasm2.Expose("VerifyPassword", storage.VerifyPassword)

	// storage/purge.go
	wasm2.Expose("Purge", storage.Purge)

	// utils/array.go
	wasm2.Expose("Uint8ArrayToBase64", utils.Uint8ArrayToBase64)
	wasm2.Expose("Base64ToUint8Array", utils.Base64ToUint8Array)
	wasm2.Expose("Uint8ArrayEquals", utils.Uint8ArrayEquals)

	// wasm/backup.go
	wasm2.Expose("NewCmixFromBackup", bindings.NewCmixFromBackup)
	wasm2.Expose("InitializeBackup", wasm.InitializeBackup)
	wasm2.Expose("ResumeBackup", wasm.ResumeBackup)

	// wasm/channels.go
	wasm2.Expose("GenerateChannelIdentity",
		bindings.GenerateChannelIdentity)
	wasm2.Expose("ConstructIdentity", wasm.ConstructIdentity)
	wasm2.Expose("ImportPrivateIdentity", bindings.ImportPrivateIdentity)
	wasm2.Expose("GetPublicChannelIdentity", bindings.GetPublicChannelIdentity)
	wasm2.Expose("GetPublicChannelIdentityFromPrivate",
		bindings.GetPublicChannelIdentityFromPrivate)
	wasm2.Expose("NewChannelsManager", wasm.NewChannelsManager)
	wasm2.Expose("LoadChannelsManager", wasm.LoadChannelsManager)
	wasm2.Expose("NewChannelsManagerWithIndexedDb",
		wasm.NewChannelsManagerWithIndexedDb)
	wasm2.Expose("LoadChannelsManagerWithIndexedDb",
		wasm.LoadChannelsManagerWithIndexedDb)
	wasm2.Expose("LoadChannelsManagerWithIndexedDbUnsafe",
		wasm.LoadChannelsManagerWithIndexedDbUnsafe)
	wasm2.Expose("NewChannelsManagerWithIndexedDbUnsafe",
		wasm.NewChannelsManagerWithIndexedDbUnsafe)
	wasm2.Expose("DecodePublicURL", bindings.DecodePublicURL)
	wasm2.Expose("DecodePrivateURL", bindings.DecodePrivateURL)
	wasm2.Expose("GetChannelJSON", bindings.GetChannelJSON)
	wasm2.Expose("GetChannelInfo", bindings.GetChannelInfo)
	wasm2.Expose("GetShareUrlType", bindings.GetShareUrlType)
	wasm2.Expose("ValidForever", bindings.ValidForever)
	wasm2.Expose("IsNicknameValid", bindings.IsNicknameValid)
	wasm2.Expose("NewChannelsDatabaseCipher",
		bindings.NewChannelsDatabaseCipher)

	// wasm/dm.go
	wasm2.Expose("NewDMClient", wasm.NewDMClient)
	wasm2.Expose("NewDMClientWithIndexedDb",
		wasm.NewDMClientWithIndexedDb)
	wasm2.Expose("NewDMClientWithIndexedDbUnsafe",
		wasm.NewDMClientWithIndexedDbUnsafe)
	wasm2.Expose("NewDMsDatabaseCipher",
		wasm.NewDMsDatabaseCipher)

	// wasm/cmix.go
	wasm2.Expose("NewCmix", wasm.NewCmix)
	wasm2.Expose("LoadCmix", wasm.LoadCmix)

	// wasm/delivery.go
	wasm2.Expose("SetDashboardURL", bindings.SetDashboardURL)

	// wasm/dummy.go
	wasm2.Expose("NewDummyTrafficManager",
		bindings.NewDummyTrafficManager)

	// wasm/e2e.go
	wasm2.Expose("Login", wasm.Login)
	wasm2.Expose("LoginEphemeral", wasm.LoginEphemeral)

	// wasm/emoji.go
	wasm2.Expose("SupportedEmojis", wasm.SupportedEmojis)
	wasm2.Expose("ValidateReaction", wasm.ValidateReaction)

	// wasm/errors.go
	wasm2.Expose("CreateUserFriendlyErrorMessage",
		wasm.CreateUserFriendlyErrorMessage)
	wasm2.Expose("UpdateCommonErrors",
		wasm.UpdateCommonErrors)

	// wasm/fileTransfer.go
	wasm2.Expose("InitFileTransfer", wasm.InitFileTransfer)

	// wasm/group.go
	wasm2.Expose("NewGroupChat", wasm.NewGroupChat)
	wasm2.Expose("DeserializeGroup", wasm.DeserializeGroup)

	// wasm/identity.go
	wasm2.Expose("StoreReceptionIdentity",
		bindings.StoreReceptionIdentity)
	wasm2.Expose("LoadReceptionIdentity",
		bindings.LoadReceptionIdentity)
	wasm2.Expose("GetContactFromReceptionIdentity",
		wasm.GetContactFromReceptionIdentity)
	wasm2.Expose("GetIDFromContact",
		bindings.GetIDFromContact)
	wasm2.Expose("GetPubkeyFromContact",
		bindings.GetPubkeyFromContact)
	wasm2.Expose("SetFactsOnContact",
		bindings.SetFactsOnContact)
	wasm2.Expose("GetFactsFromContact",
		bindings.GetFactsFromContact)

	// wasm/logging.go
	wasm2.Expose("LogLevel", wasm.LogLevel)
	wasm2.Expose("RegisterLogWriter", wasm.RegisterLogWriter)
	wasm2.Expose("EnableGrpcLogs", wasm.EnableGrpcLogs)

	// wasm/ndf.go
	wasm2.Expose("DownloadAndVerifySignedNdfWithUrl",
		wasm.DownloadAndVerifySignedNdfWithUrl)

	// wasm/params.go
	wasm2.Expose("GetDefaultCMixParams",
		wasm.GetDefaultCMixParams)
	wasm2.Expose("GetDefaultE2EParams",
		wasm.GetDefaultE2EParams)
	wasm2.Expose("GetDefaultFileTransferParams",
		wasm.GetDefaultFileTransferParams)
	wasm2.Expose("GetDefaultSingleUseParams",
		wasm.GetDefaultSingleUseParams)
	wasm2.Expose("GetDefaultE2eFileTransferParams",
		wasm.GetDefaultE2eFileTransferParams)

	// wasm/restlike.go
	wasm2.Expose("RestlikeRequest", wasm.RestlikeRequest)
	wasm2.Expose("RestlikeRequestAuth", wasm.RestlikeRequestAuth)

	// wasm/restlikeSingle.go
	wasm2.Expose("RequestRestLike",
		wasm.RequestRestLike)
	wasm2.Expose("AsyncRequestRestLike",
		wasm.AsyncRequestRestLike)

	// wasm/secrets.go
	wasm2.Expose("GenerateSecret", wasm.GenerateSecret)

	// wasm/single.go
	wasm2.Expose("TransmitSingleUse", wasm.TransmitSingleUse)
	wasm2.Expose("Listen", wasm.Listen)

	// wasm/timeNow.go
	wasm2.Expose("SetTimeSource", wasm.SetTimeSource)
	wasm2.Expose("SetOffset", wasm.SetOffset)

	// wasm/ud.go
	wasm2.Expose("NewOrLoadUd", wasm.NewOrLoadUd)
	wasm2.Expose("NewUdManagerFromBackup",
		wasm.NewUdManagerFromBackup)
	wasm2.Expose("LookupUD", wasm.LookupUD)
	wasm2.Expose("SearchUD", wasm.SearchUD)

	// wasm/version.go
	wasm2.Expose("GetVersion", wasm.GetVersion)
	wasm2.Expose("GetClientVersion", wasm.GetClientVersion)
	wasm2.Expose("GetClientGitVersion", wasm.GetClientGitVersion)
	wasm2.Expose("GetClientDependencies", wasm.GetClientDependencies)
	wasm2.Expose("GetWasmSemanticVersion", wasm.GetWasmSemanticVersion)
	wasm2.Expose("GetXXDKSemanticVersion", wasm.GetXXDKSemanticVersion)

	wasm2.Ready()
	<-make(chan bool)
	os.Exit(0)
}
