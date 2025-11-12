import type { XXDKUtils } from './types';
import { logFileWorkerPath, stateIndexedDbWorkerPath } from './paths';

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
  const NewChannelsManagerWithIndexedDb = wasmRaw.NewChannelsManagerWithIndexedDb as any;
  const NewCmix = wasmRaw.NewCmix as any;
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
        w.postMessage(JSON.stringify({ tag: 'GetFileExt' }))
      });
    };

    window.logger = logger
  }

  // GenericKeyValue implementation that checks for worker on every call
  // Falls back to direct IndexedDB if state worker is not available
  const createGenericKeyValue = async (storageDir: string) => {
    // Lazy state model initialization
    let stateModel: any = null;
    let stateModelPath: string | null = null;

    // Fallback IndexedDB setup
    const dbName = `xxdk-kv-${storageDir}`;
    const storeName = 'genericKV';
    let db: IDBDatabase | null = null;

    const initFallbackDb = async (): Promise<IDBDatabase> => {
      if (db) return db;
      return new Promise((resolve, reject) => {
        const request = indexedDB.open(dbName, 1);
        request.onerror = () => reject(request.error);
        request.onsuccess = () => {
          db = request.result;
          resolve(request.result);
        };
        request.onupgradeneeded = (event) => {
          const database = (event.target as IDBOpenDBRequest).result;
          if (!database.objectStoreNames.contains(storeName)) {
            database.createObjectStore(storeName);
          }
        };
      });
    };

    // Check for worker availability on each call
    const getStateModel = async () => {
      const wasmRawCheck = window as any;
      if (wasmRawCheck.NewState && typeof wasmRawCheck.NewState === 'function') {
        if (!stateModel || !stateModelPath) {
          console.log('[XXDK] Initializing state worker for GenericKeyValue');
          const workerPath = await stateIndexedDbWorkerPath();
          stateModelPath = workerPath.toString();
          stateModel = await wasmRawCheck.NewState(storageDir, stateModelPath);
        }
        return stateModel;
      }
      return null;
    };

    return {
      Get: async (key: string): Promise<Uint8Array> => {
        const model = await getStateModel();
        if (model) {
          return await model.Get(key);
        } else {
          // Fallback to direct IndexedDB
          const database = await initFallbackDb();
          return new Promise((resolve, reject) => {
            const transaction = database.transaction([storeName], 'readonly');
            const store = transaction.objectStore(storeName);
            const request = store.get(key);
            request.onerror = () => reject(request.error);
            request.onsuccess = () => {
              if (request.result === undefined) {
                reject(new Error(`Key not found: ${key}`));
              } else {
                resolve(request.result);
              }
            };
          });
        }
      },
      Set: async (key: string, value: Uint8Array): Promise<void> => {
        const model = await getStateModel();
        if (model) {
          return await model.Set(key, value);
        } else {
          // Fallback to direct IndexedDB
          const database = await initFallbackDb();
          return new Promise((resolve, reject) => {
            const transaction = database.transaction([storeName], 'readwrite');
            const store = transaction.objectStore(storeName);
            const request = store.put(value, key);
            request.onerror = () => reject(request.error);
            request.onsuccess = () => resolve();
          });
        }
      },
      Delete: async (key: string): Promise<void> => {
        const model = await getStateModel();
        if (model) {
          return await model.Delete(key);
        } else {
          // Fallback to direct IndexedDB
          const database = await initFallbackDb();
          return new Promise((resolve, reject) => {
            const transaction = database.transaction([storeName], 'readwrite');
            const store = transaction.objectStore(storeName);
            const request = store.delete(key);
            request.onerror = () => reject(request.error);
            request.onsuccess = () => resolve();
          });
        }
      },
      Keys: async (): Promise<Uint8Array> => {
        const model = await getStateModel();
        if (model) {
          return await model.Keys();
        } else {
          // Fallback to direct IndexedDB
          const database = await initFallbackDb();
          return new Promise((resolve, reject) => {
            const transaction = database.transaction([storeName], 'readonly');
            const store = transaction.objectStore(storeName);
            const request = store.getAllKeys();
            request.onerror = () => reject(request.error);
            request.onsuccess = () => {
              const keys = request.result as string[];
              const encoder = new TextEncoder();
              resolve(encoder.encode(JSON.stringify(keys)));
            };
          });
        }
      }
    };
  };

  // Wrapper for NewCmix that automatically creates and uses GenericKeyValue
  const NewCmixWrapper = async (
    ndf: string,
    storageDir: string,
    password: Uint8Array,
    registrationCode: string
  ) => {
    console.log('[XXDK] NewCmix called - creating GenericKeyValue for:', storageDir);
    const kv = await createGenericKeyValue(storageDir);
    return NewCmix(kv, ndf, storageDir, password, registrationCode);
  };

  // Wrapper for LoadCmix that automatically creates and uses GenericKeyValue
  const LoadCmixWrapper = async (
    storageDirectory: string,
    password: Uint8Array,
    cmixParams: Uint8Array
  ) => {
    console.log('[XXDK] LoadCmix called - creating GenericKeyValue for:', storageDirectory);
    const kv = await createGenericKeyValue(storageDirectory);
    return LoadCmix(kv, storageDirectory, password, cmixParams);
  };

  // Return WASM functions (SafeFunc automatically provides Promises)
  // Note: NewCmix and LoadCmix are wrapped to automatically use GenericKeyValue
  xxdkUtils({
    NewCmix: NewCmixWrapper,
    LoadCmix: LoadCmixWrapper,
    // Also expose the raw KV-based functions for advanced users
    NewCmixWithKV: NewCmix,
    LoadCmixWithKV: LoadCmix,
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
