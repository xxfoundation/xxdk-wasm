import type { XXDKUtils } from './types';

declare global {
  interface Window extends XXDKUtils {}
}

// @ts-ignore
// import makeWasm from '../xxdk.wasm';

export const loadUtils = () => new Promise<XXDKUtils>(async (res) => {
    const Go = require('../wasm_exec');
    const go = new Go();
    let mod, inst;
    WebAssembly.instantiateStreaming(fetch("../xxdk.wasm"), go.importObject).then(async (result) => {
        mod = result.module;
        inst = result.instance;
        await go.run(inst);
        inst = await WebAssembly.instantiate(mod, go.importObject); // reset instance
    }).catch((err) => {
        console.error(err);
    });

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
