import type { XXDKUtils } from './types';

import '../wasm_exec.js';

const xxdkWasm: URL = require('../xxdk.wasm');

declare global {
  interface Window extends XXDKUtils {
            Go: any;
  }
}


var xxdk_base_path: string = "";
export function setPath(newPath: string) {
        xxdk_base_path = newPath;
}

export const loadWasm = () => new Promise<void>(async () => {
    const xxdk_wasm_path = xxdk_base_path + xxdkWasm;
    // if (typeof window == "undefined") {
    const go = new window!.Go();
    console.log(go);
    console.log(xxdk_wasm_path);
    console.log("IMPORT");
    console.log(go.importObject);
    let stream = await WebAssembly.instantiateStreaming(fetch(xxdk_wasm_path), go.importObject);
    go.run(stream.instance);
});


export const loadUtils = () => new Promise<XXDKUtils>(async (res) => {

  const {
    Base64ToUint8Array,
    Crash,
    ConstructIdentity,
    DecodePrivateURL,
    DecodePublicURL,
    GenerateChannelIdentity,
    GetChannelInfo,
    GetChannelJSON,
    GetClientVersion,
    getCrashedLogFile,
    GetDefaultCMixParams,
    GetLogger,
    GetOrInitPassword,
    GetPublicChannelIdentityFromPrivate,
    GetShareUrlType,
    GetVersion,
    GetWasmSemanticVersion,
    ImportPrivateIdentity,
    IsNicknameValid,
    LoadChannelsManagerWithIndexedDb,
    LoadCmix,
    LogLevel,
    NewChannelsDatabaseCipher,
    NewChannelsManagerWithIndexedDb,
    NewCmix,
    NewDMClientWithIndexedDb,
    NewDMsDatabaseCipher,
    NewDummyTrafficManager,
    Purge,
    ValidForever,
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  } = (window) || {};

  res({
    Base64ToUint8Array,
    Crash,
    ConstructIdentity,
    DecodePrivateURL,
    DecodePublicURL,
    GenerateChannelIdentity,
    GetChannelInfo,
    GetChannelJSON,
    GetClientVersion,
    getCrashedLogFile,
    GetDefaultCMixParams,
    GetLogger,
    GetOrInitPassword,
    GetPublicChannelIdentityFromPrivate,
    GetShareUrlType,
    GetVersion,
    GetWasmSemanticVersion,
    ImportPrivateIdentity,
    IsNicknameValid,
    LoadChannelsManagerWithIndexedDb,
    LoadCmix,
    LogLevel,
    NewChannelsDatabaseCipher,
    NewChannelsManagerWithIndexedDb,
    NewCmix,
    NewDMClientWithIndexedDb,
    NewDMsDatabaseCipher,
    NewDummyTrafficManager,
    Purge,
    ValidForever,
  })
})

export * from './types';
