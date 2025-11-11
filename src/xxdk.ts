import type { XXDKUtils } from './types';
import { logFileWorkerPath } from './paths';

const xxdkWasm: URL = require('../assets/wasm/xxdk.wasm');

type Logger = {
  StopLogging: () => void,
  GetFile: () => Promise<string>,
  Threshold: () => number,
  MaxSize: () => number,
  Size: () => Promise<number>,
  Worker: () => Worker,
};

declare global {
  interface Window extends XXDKUtils {
      onWasmInitialized: () => void;
      Crash: () => void;
      GetLogger: () => Logger;
      logger?: Logger;
      getCrashedLogFile: () => Promise<string>;
      Go: any;
  }
}

export const InitXXDK = () => new Promise<XXDKUtils>(async (xxdkUtils) => {
  await import('../wasm_exec.js');
  const isReady = new Promise<void>((resolve) => {
    window!.onWasmInitialized = resolve;
  });

  const xxdkWasmPath = new URL(window!.xxdkBasePath.toString() + xxdkWasm.toString());
  console.log("Fetching xxdkWASM: " + xxdkWasmPath);
  console.log("Fetching xxdkWASM base: " + window!.xxdkBasePath);
  console.log("Fetching xxdkWASM path: " + xxdkWasm);

  const logWorker = await logFileWorkerPath();
  console.log("Got logworkerURL: " + logWorker);
  let go = new window!.Go();
  go.argv = [
    '--logLevel=1',
    '--fileLogLevel=1',
    '--workerScriptURL=' + logWorker,
  ]

  // Debug: Check main WASM fetch
  console.log("[XXDK] Fetching main WASM from:", xxdkWasmPath.toString());
  const mainWasmResponse = await fetch(xxdkWasmPath);
  console.log("[XXDK] Main WASM fetch status:", mainWasmResponse.status);
  console.log("[XXDK] Main WASM content-type:", mainWasmResponse.headers.get('content-type'));

  if (!mainWasmResponse.ok) {
    const text = await mainWasmResponse.text();
    console.error("[XXDK] Main WASM fetch failed. Response body:", text.substring(0, 500));
    throw new Error(`Failed to fetch main WASM: ${mainWasmResponse.status} ${mainWasmResponse.statusText}`);
  }

  let stream = await WebAssembly?.instantiateStreaming(
    mainWasmResponse, go.importObject);
  go.run(stream.instance);
  await isReady;

  // Get functions from WASM
  // All functions wrapped in SafeFunc return Promises
  const wasmRaw = window as any;

  // Functions that return Promises (wrapped in SafeFunc in Go)
  // Cast to any to avoid type signature mismatches - the actual types are defined in XXDKUtils
  const Base64ToUint8Array = wasmRaw.Base64ToUint8Array as any;
  const ConstructIdentity = wasmRaw.ConstructIdentity as any;
  const DecodePrivateURL = wasmRaw.DecodePrivateURL as any;
  const DecodePublicURL = wasmRaw.DecodePublicURL as any;
  const GenerateChannelIdentity = wasmRaw.GenerateChannelIdentity as any;
  const GetChannelInfo = wasmRaw.GetChannelInfo as any;
  const GetChannelJSON = wasmRaw.GetChannelJSON as any;
  const GetClientVersion = wasmRaw.GetClientVersion as any;
  const GetOrInitPassword = wasmRaw.GetOrInitPassword as any;
  const GetPublicChannelIdentityFromPrivate = wasmRaw.GetPublicChannelIdentityFromPrivate as any;
  const GetShareUrlType = wasmRaw.GetShareUrlType as any;
  const GetVersion = wasmRaw.GetVersion as any;
  const GetWasmSemanticVersion = wasmRaw.GetWasmSemanticVersion as any;
  const ImportPrivateIdentity = wasmRaw.ImportPrivateIdentity as any;
  const IsNicknameValid = wasmRaw.IsNicknameValid as any;
  const LoadChannelsManagerWithIndexedDb = wasmRaw.LoadChannelsManagerWithIndexedDb as any;
  const LoadCmix = wasmRaw.LoadCmix as any;
  const LoadNotifications = wasmRaw.LoadNotifications as any;
  const LoadNotificationsDummy = wasmRaw.LoadNotificationsDummy as any;
  const LoadSynchronizedCmix = wasmRaw.LoadSynchronizedCmix as any;
  const NewChannelsManagerWithIndexedDb = wasmRaw.NewChannelsManagerWithIndexedDb as any;
  const NewCmix = wasmRaw.NewCmix as any;
  const NewDatabaseCipher = wasmRaw.NewDatabaseCipher as any;
  const NewDMClientWithIndexedDb = wasmRaw.NewDMClientWithIndexedDb as any;
  const NewDummyTrafficManager = wasmRaw.NewDummyTrafficManager as any;
  const NewSynchronizedCmix = wasmRaw.NewSynchronizedCmix as any;
  const Purge = wasmRaw.Purge as any;
  const RPCSend = wasmRaw.RPCSend as any;
  const ValidForever = wasmRaw.ValidForever as any;

  // Parameter getters are truly synchronous - not wrapped in SafeFunc
  const GetDefaultCMixParams = wasmRaw.GetDefaultCMixParams as any;
  const GetDefaultE2EParams = wasmRaw.GetDefaultE2EParams as any;
  const GetDefaultE2eFileTransferParams = wasmRaw.GetDefaultE2eFileTransferParams as any;
  const GetDefaultFileTransferParams = wasmRaw.GetDefaultFileTransferParams as any;
  const GetDefaultSingleUseParams = wasmRaw.GetDefaultSingleUseParams as any;

  const { GetLogger } = window;
  if(GetLogger) {
    const logger = GetLogger()

    // Get the actual Worker object from the log file object
    const w = logger.Worker()

    window.getCrashedLogFile = () => {
      return new Promise((resolve) => {
        w.addEventListener('message', ev => {
          resolve(atob(JSON.parse(ev.data).data))
        })
        w.postMessage(JSON.stringify({ tag: 'GetFileExt' }))
      });
    };

    window.logger = logger
  }

  // Return WASM functions (SafeFunc automatically provides Promises)
  xxdkUtils({
    NewCmix,
    NewSynchronizedCmix,
    LoadCmix,
    LoadNotifications,
    LoadNotificationsDummy,
    LoadSynchronizedCmix,
    GetChannelInfo,
    GenerateChannelIdentity,
    GetDefaultCMixParams,
    GetDefaultE2EParams,
    GetDefaultE2eFileTransferParams,
    GetDefaultFileTransferParams,
    GetDefaultSingleUseParams,
    NewChannelsManagerWithIndexedDb,
    Base64ToUint8Array,
    LoadChannelsManagerWithIndexedDb,
    GetPublicChannelIdentityFromPrivate,
    IsNicknameValid,
    GetShareUrlType,
    GetVersion,
    GetClientVersion,
    GetOrInitPassword,
    GetWasmSemanticVersion,
    ImportPrivateIdentity,
    ConstructIdentity,
    DecodePrivateURL,
    DecodePublicURL,
    GetChannelJSON,
    NewDMClientWithIndexedDb,
    NewDatabaseCipher,
    NewDummyTrafficManager,
    Purge,
    ValidForever,
    RPCSend
  });
});
