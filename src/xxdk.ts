import type { XXDKUtils } from './types';
import { logFileWorkerPath } from './paths';

import DefaultNdf from './ndf.json'

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

  let stream = await WebAssembly?.instantiateStreaming(
    fetch(xxdkWasmPath), go.importObject);
  go.run(stream.instance);
  await isReady;

  const {
    Base64ToUint8Array,
    ConstructIdentity,
    DecodePrivateURL,
    DecodePublicURL,
    GenerateChannelIdentity,
    GetChannelInfo,
    GetChannelJSON,
    GetClientVersion,
    GetDefaultCMixParams,
    GetOrInitPassword,
    GetPublicChannelIdentityFromPrivate,
    GetShareUrlType,
    GetVersion,
    GetWasmSemanticVersion,
    ImportPrivateIdentity,
    IsNicknameValid,
    LoadChannelsManagerWithIndexedDb,
    LoadCmix,
    LoadNotifications,
    LoadNotificationsDummy,
    LoadSynchronizedCmix,
    NewChannelsManagerWithIndexedDb,
    NewCmix,
    NewDMClientWithIndexedDb,
    NewDatabaseCipher,
    NewDummyTrafficManager,
    NewSynchronizedCmix,
    Purge,
    ValidForever,
    RPCSend,
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  } = (window as any) || {};

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

export const GetDefaultNDF = (): String => {
  return JSON.stringify(DefaultNdf);
}