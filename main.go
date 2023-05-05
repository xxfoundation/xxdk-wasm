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
	"github.com/spf13/cobra"
	"os"
	"syscall/js"

	jww "github.com/spf13/jwalterweatherman"

	"gitlab.com/elixxir/xxdk-wasm/logging"
	"gitlab.com/elixxir/xxdk-wasm/storage"
	"gitlab.com/elixxir/xxdk-wasm/utils"
	"gitlab.com/elixxir/xxdk-wasm/wasm"
)

func main() {
	// Set to os.Args because the default is os.Args[1:] and in WASM, args start
	// at 0, not 1.
	wasmCmd.SetArgs(os.Args)

	err := wasmCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var wasmCmd = &cobra.Command{
	Use:     "xxdk-wasm",
	Short:   "WebAssembly bindings for xxDK.",
	Example: "const go = new Go();\ngo.argv = [\"--logLevel=1\"]",
	Run: func(cmd *cobra.Command, args []string) {
		// Start logger first to capture all logging events
		err := logging.EnableLogging(
			logLevel, fileLogLevel, maxLogFileSize, workerScriptURL, workerName)
		if err != nil {
			fmt.Printf("Failed to intialize logging: %+v", err)
			os.Exit(1)
		}

		// Check that the WASM binary version is correct
		err = storage.CheckAndStoreVersions()
		if err != nil {
			jww.FATAL.Panicf("WASM binary version error: %+v", err)
		}

		// Enable all top level bindings functions
		setGlobals()

		// Indicate to the Javascript caller that the WASM is ready by resolving
		// a promise created by the caller, as shown below:
		//
		//  let isReady = new Promise((resolve) => {
		//    window.onWasmInitialized = resolve;
		//  });
		//
		//  const go = new Go();
		//  go.run(result.instance);
		//  await isReady;
		//
		// Source: https://github.com/golang/go/issues/49710#issuecomment-986484758
		js.Global().Get("onWasmInitialized").Invoke()

		<-make(chan bool)
		os.Exit(0)
	},
}

// setGlobals enables all global functions to be accessible to Javascript.
func setGlobals() {
	jww.INFO.Printf("Starting xxDK WebAssembly bindings.")

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
	js.Global().Set("DecodePublicURL", js.FuncOf(wasm.DecodePublicURL))
	js.Global().Set("DecodePrivateURL", js.FuncOf(wasm.DecodePrivateURL))
	js.Global().Set("GetChannelJSON", js.FuncOf(wasm.GetChannelJSON))
	js.Global().Set("GetChannelInfo", js.FuncOf(wasm.GetChannelInfo))
	js.Global().Set("GetShareUrlType", js.FuncOf(wasm.GetShareUrlType))
	js.Global().Set("ValidForever", js.FuncOf(wasm.ValidForever))
	js.Global().Set("IsNicknameValid", js.FuncOf(wasm.IsNicknameValid))
	js.Global().Set("GetNoMessageErr", js.FuncOf(wasm.GetNoMessageErr))
	js.Global().Set("CheckNoMessageErr", js.FuncOf(wasm.CheckNoMessageErr))
	js.Global().Set("NewChannelsDatabaseCipher",
		js.FuncOf(wasm.NewChannelsDatabaseCipher))

	// wasm/dm.go
	js.Global().Set("InitChannelsFileTransfer",
		js.FuncOf(wasm.InitChannelsFileTransfer))

	// wasm/dm.go
	js.Global().Set("NewDMClient", js.FuncOf(wasm.NewDMClient))
	js.Global().Set("NewDMClientWithIndexedDb",
		js.FuncOf(wasm.NewDMClientWithIndexedDb))
	js.Global().Set("NewDMClientWithIndexedDbUnsafe",
		js.FuncOf(wasm.NewDMClientWithIndexedDbUnsafe))
	js.Global().Set("NewDMsDatabaseCipher",
		js.FuncOf(wasm.NewDMsDatabaseCipher))

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

	// wasm/emoji.go
	js.Global().Set("SupportedEmojis", js.FuncOf(wasm.SupportedEmojis))
	js.Global().Set("SupportedEmojisMap", js.FuncOf(wasm.SupportedEmojisMap))
	js.Global().Set("ValidateReaction", js.FuncOf(wasm.ValidateReaction))

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
	js.Global().Set("GetWasmSemanticVersion", js.FuncOf(wasm.GetWasmSemanticVersion))
	js.Global().Set("GetXXDKSemanticVersion", js.FuncOf(wasm.GetXXDKSemanticVersion))
}

var (
	logLevel, fileLogLevel      jww.Threshold
	maxLogFileSize              int
	workerScriptURL, workerName string
)

func init() {
	// Initialize all startup flags
	wasmCmd.Flags().IntVarP((*int)(&logLevel), "logLevel", "l", 2,
		"Sets the log level output when outputting to the Javascript console. "+
			"0 = TRACE, 1 = DEBUG, 2 = INFO, 3 = WARN, 4 = ERROR, "+
			"5 = CRITICAL, 6 = FATAL, -1 = disabled.")
	wasmCmd.Flags().IntVarP((*int)(&fileLogLevel), "fileLogLevel", "m", -1,
		"The log level when outputting to the file buffer. "+
			"0 = TRACE, 1 = DEBUG, 2 = INFO, 3 = WARN, 4 = ERROR, "+
			"5 = CRITICAL, 6 = FATAL, -1 = disabled.")
	wasmCmd.Flags().IntVarP(&maxLogFileSize, "maxLogFileSize", "s", 5_000_000,
		"Max file size, in bytes, for the file buffer before it rolls over "+
			"and starts overwriting the oldest entries.")
	wasmCmd.Flags().StringVarP(&workerScriptURL, "workerScriptURL", "w", "",
		"URL to the script that executes the worker. If set, it enables the "+
			"saving of log file to buffer in Worker instead of in the local "+
			"thread. This allows logging to be available after the main WASM "+
			"thread crashes.")
	wasmCmd.Flags().StringVar(&workerName, "workerName", "xxdkLogFileWorker",
		"Name of the logger worker.")
}
