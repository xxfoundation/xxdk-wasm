import type { XXDKUtils } from './types';
import { logFileWorkerPath, kvWorkerPath } from './paths';
import { createKVStore, setKVWorkerUrl, cleanupKVWorkers } from './kv';

/**
 * Cleans up all workers (Go-managed and TypeScript-managed).
 * This should be called before page unload or when shutting down the SDK.
 *
 * Workers cleaned up:
 * - Go-managed: Channels, DM, other WASM workers (via StopWorkers)
 * - TypeScript-managed: KV Worker (via cleanupKVWorkers)
 * - Logger worker (via StopLogging)
 */
export function cleanupAllWorkers(): void {
  console.log('[XXDK] Cleaning up all workers...');

  // Stop all Go-managed workers (channels, DM, etc.)
  if (typeof (window as any).StopWorkers === 'function') {
    try {
      (window as any).StopWorkers();
      console.log('[XXDK] Go workers stopped');
    } catch (e) {
      console.warn('[XXDK] Failed to stop Go workers:', e);
    }
  }

  // Stop TypeScript-managed KV worker
  cleanupKVWorkers();

  // Stop logger worker
  if (window.logger) {
    try {
      window.logger.StopLogging();
      console.log('[XXDK] Logger stopped');
    } catch (e) {
      console.warn('[XXDK] Failed to stop logger:', e);
    }
  }

  console.log('[XXDK] All workers cleaned up');
}

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

// Guard to prevent duplicate WASM initialization (e.g., from React re-renders)
let xxdkInitialized = false;
let xxdkInitPromise: Promise<XXDKUtils> | null = null;

export const InitXXDK = () => {
  // Return existing promise if already initializing or initialized
  if (xxdkInitPromise) {
    console.log('[XXDK] Initialization already in progress or complete, returning existing promise');
    return xxdkInitPromise;
  }

  xxdkInitPromise = new Promise<XXDKUtils>(async (resolve, reject) => {
    if (xxdkInitialized) {
      console.warn('[XXDK] Already initialized, skipping duplicate initialization');
      return;
    }
    xxdkInitialized = true;

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

  console.log("[XXDK] DEBUG: Starting WebAssembly.instantiateStreaming...");
  let stream = await WebAssembly?.instantiateStreaming(
    mainWasmResponse, go.importObject);
  console.log("[XXDK] DEBUG: WebAssembly instantiated, calling go.run()...");
  go.run(stream.instance);
  console.log("[XXDK] DEBUG: go.run() called (async), awaiting isReady (onWasmInitialized callback)...");

  // Add a timeout to detect if onWasmInitialized is never called
  const timeoutPromise = new Promise<void>((_, reject) => {
    setTimeout(() => {
      reject(new Error('[XXDK] TIMEOUT: onWasmInitialized was never called after 30 seconds. Go WASM may have crashed or hung during initialization.'));
    }, 30000);
  });

  try {
    await Promise.race([isReady, timeoutPromise]);
    console.log("[XXDK] DEBUG: onWasmInitialized callback received, WASM is ready!");
  } catch (err) {
    console.error("[XXDK] DEBUG: Error waiting for WASM initialization:", err);
    throw err;
  }

  // Note: KV worker registration with Go happens inside createKVStore() when the worker is created

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
  const LoadCmixWithKV = wasmRaw.LoadCmixWithKV as any;
  const LoadNotifications = wasmRaw.LoadNotifications as any;
  const LoadNotificationsDummy = wasmRaw.LoadNotificationsDummy as any;
  const NewChannelsManagerWithIndexedDb = wasmRaw.NewChannelsManagerWithIndexedDb as any;
  const NewCmix = wasmRaw.NewCmix as any;
  const NewCmixWithKV = wasmRaw.NewCmixWithKV as any;
  const NewDatabaseCipher = wasmRaw.NewDatabaseCipher as any;
  const NewDMClientWithIndexedDb = wasmRaw.NewDMClientWithIndexedDb as any;
  const NewDummyTrafficManager = wasmRaw.NewDummyTrafficManager as any;
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
        const message = {
          tag: 'GetFileExt',
          id: 0,
          response: false,
          data: ''
        };
        const messageBytes = new TextEncoder().encode(JSON.stringify(message));
        w.postMessage(messageBytes);
      });
    };

    window.logger = logger
  }

  // Initialize KV Worker URL for persistent IndexedDB storage
  // This must be called before any createKVStore calls
  try {
    const kvWorkerUrl = await kvWorkerPath();
    setKVWorkerUrl(kvWorkerUrl);
    console.log('[XXDK] KV Worker URL initialized:', kvWorkerUrl);
  } catch (err) {
    console.warn('[XXDK] Failed to initialize KV Worker, falling back to direct IndexedDB:', err);
  }

  // Ensure KV Worker is initialized for the given storage directory
  // This creates the KV Worker and registers it with Go via SetKVWorkerManager
  const ensureKVWorker = async (storageDir: string): Promise<void> => {
    // createKVStore initializes the worker and calls SetKVWorkerManager
    await createKVStore(storageDir);
    console.log('[XXDK] KV Worker initialized for:', storageDir);
  };

  // Wrapper for NewCmix that ensures KV Worker is ready
  const NewCmixWrapper = async (
    ndf: string,
    storageDir: string,
    password: Uint8Array,
    registrationCode: string
  ) => {
    console.log('[XXDK] NewCmixWrapper called with storageDir:', storageDir);
    await ensureKVWorker(storageDir);
    // kvPath is passed but Go uses the global store set by SetKVWorkerManager
    const result = await NewCmixWithKV(storageDir, ndf, storageDir, password, registrationCode);
    console.log('[XXDK] NewCmixWithKV returned:', result ? 'success' : 'null/undefined');
    return result;
  };

  // Wrapper for LoadCmix that ensures KV Worker is ready
  const LoadCmixWrapper = async (
    storageDirectory: string,
    password: Uint8Array,
    cmixParams: Uint8Array
  ) => {
    console.log('[XXDK] LoadCmixWrapper called with storageDir:', storageDirectory);
    await ensureKVWorker(storageDirectory);
    // kvPath is passed but Go uses the global store set by SetKVWorkerManager
    const result = await LoadCmixWithKV(storageDirectory, storageDirectory, password, cmixParams);
    console.log('[XXDK] LoadCmixWithKV returned:', result ? 'success' : 'null/undefined');
    return result;
  };

  // Register cleanup handler to terminate workers on page unload/reload
  // This ensures IndexedDB connections are released properly
  window.addEventListener('beforeunload', () => {
    console.log('[XXDK] Page unloading...');
    cleanupAllWorkers();
  });
  console.log('[XXDK] Registered beforeunload cleanup handler');

  // Return WASM functions (SafeFunc automatically provides Promises)
  // Note: NewCmix and LoadCmix are wrapped to automatically use GenericKeyValue
  console.log('[XXDK] Resolving InitXXDK promise with utils');
  resolve({
    // Wrappers that auto-create and manage GenericKeyValue (recommended)
    NewCmix: NewCmixWrapper,
    LoadCmix: LoadCmixWrapper,
    // Raw functions for advanced users who want to manage their own KV
    NewCmixWithKV: NewCmixWithKV,
    LoadCmixWithKV: LoadCmixWithKV,
    LoadNotifications,
    LoadNotificationsDummy,
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

return xxdkInitPromise;
};
