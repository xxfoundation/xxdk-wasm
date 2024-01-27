import type { XXDKUtils } from './types';

declare global {
  interface Window extends XXDKUtils {}
}

import './wasm_exec';
// @ts-ignore
import makeWasm from './xxdk.wasm';

export const loadUtils = () => new Promise<XXDKUtils>(async (res) => {
  const go = new (window as any).Go();
  const result = await makeWasm(go.importObject);

  go.run(result.instance);

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
