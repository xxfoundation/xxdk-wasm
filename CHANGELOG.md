# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.4.0] - 2025-11-11

- Updated to use Go WASM's `SafeFunc` wrapper for error handling
- Functions now return Promises for consistent error handling

### Breaking Changes

#### Functions Now Requiring `await` or `.then()`

##### Utility Functions
- `Base64ToUint8Array(base64: string)` → `Promise<Uint8Array>`
- `GetVersion()` → `Promise<string>`
- `GetClientVersion()` → `Promise<string>`
- `GetWasmSemanticVersion()` → `Promise<Uint8Array>`
- `ValidForever()` → `Promise<number>`

##### Channel Identity Functions
- `ConstructIdentity(publicKey: Uint8Array, codesetVersion: number)` → `Promise<Uint8Array>`
- `GenerateChannelIdentity(cmixId: number)` → `Promise<Uint8Array>`
- `GetPublicChannelIdentityFromPrivate(privateKey: Uint8Array)` → `Promise<Uint8Array>`
- `ImportPrivateIdentity(password: string, privateIdentity: Uint8Array)` → `Promise<Uint8Array>`

##### URL Decoding Functions
- `DecodePrivateURL(url: string, password: string)` → `Promise<string>`
- `DecodePublicURL(url: string)` → `Promise<string>`
- `GetShareUrlType(url: string)` → `Promise<any>`

##### Channel Info Functions
- `GetChannelInfo(prettyPrint: string)` → `Promise<Uint8Array>`
- `GetChannelJSON(prettyPrint: string)` → `Promise<Uint8Array>`

##### Notification Functions
- `LoadNotifications(cmixId: number)` → `Promise<Notifications>`
- `LoadNotificationsDummy(cmixId: number)` → `Promise<Notifications>`

##### Other Functions
- `IsNicknameValid(nickname: string)` → `Promise<null>`
- `NewDatabaseCipher(cmixId: number, storagePassword: Uint8Array, payloadMaximumSize: number)` → `Promise<RawCipher>`
- `NewDummyTrafficManager(cmixId: number, maximumOfMessagesPerCycle: number, durationToWaitBetweenSendsMilliseconds: number, upperBoundIntervalBetweenCyclesMilliseconds: number)` → `Promise<DummyTraffic>`
- `Purge(userPassword: string)` → `Promise<void>`

#### Functions That Remain Synchronous

These parameter getter functions remain **truly synchronous** and work correctly in synchronous contexts like `useMemo`:

- `GetDefaultCMixParams()` → `Uint8Array`
- `GetDefaultE2EParams()` → `Uint8Array`
- `GetDefaultE2eFileTransferParams()` → `Uint8Array`
- `GetDefaultFileTransferParams()` → `Uint8Array`
- `GetDefaultSingleUseParams()` → `Uint8Array`

---

## [0.3.* and earlier] - 2024

- Initial release
- WebAssembly bindings for xxDK
- TypeScript type definitions

