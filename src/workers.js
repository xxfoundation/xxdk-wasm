////////////////////////////////////////////////////////////////////////////////
// Copyright © 2024 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

// NOTE: These javascript functions are used to load web workers. They
// can't call other functions and they must use the webworker
// available APIs to function. Thats why they are duplicate code.

export function startLogFileWorker(wasm) {
  // Prevent duplicate worker initialization
  if (self.__xxdkLoggerWorkerStarted) {
    console.warn("[XXDK] logFileWorker already started, ignoring duplicate call");
    return;
  }
  self.__xxdkLoggerWorkerStarted = true;

  console.trace("[XXDK] logFileWorker loading from: " + wasm.toString());
  const isReady = new Promise((resolve) => {
    self.onWasmInitialized = resolve;
  });

  const go = new Go();
  go.argv = ['--workerType=logger'];

  // Debug: Check the fetch response before instantiateStreaming
  fetch(wasm).then(async (response) => {
    console.log("[XXDK] logFileWorker fetch response status:", response.status);
    console.log("[XXDK] logFileWorker fetch response headers:", [...response.headers.entries()]);
    console.log("[XXDK] logFileWorker fetch content-type:", response.headers.get('content-type'));

    if (!response.ok) {
      const text = await response.text();
      console.error("[XXDK] logFileWorker fetch failed. Response body:", text.substring(0, 500));
      throw new Error(`Failed to fetch WASM: ${response.status} ${response.statusText}`);
    }

    return WebAssembly.instantiateStreaming(response, go.importObject);
  }).then(async (result) => {
    go.run(result.instance);
    await isReady;
    console.info("[XXDK] logFileWorker started");
  }).catch((err) => {
    console.error("[XXDK] logFileWorker ERROR:", err);
    console.error("[XXDK] logFileWorker stack:", err.stack);
  });
}

export function startChannelsIndexedDbWorker(wasm) {
  // Prevent duplicate worker initialization
  if (self.__xxdkChannelsWorkerStarted) {
    console.warn("[XXDK] channelsIndexedDbWorker already started, ignoring duplicate call");
    return;
  }
  self.__xxdkChannelsWorkerStarted = true;

  console.trace("[XXDK] channelsIndexedDbWorker loading from: " + wasm.toString());
  const isReady = new Promise((resolve) => {
    self.onWasmInitialized = resolve;
  });
  const go = new Go();
  go.argv = [
    '--workerType=channels',
    '--logLevel=2',
    '--threadLogLevel=2',
  ]

  fetch(wasm).then(async (response) => {
    console.log("[XXDK] channelsIndexedDbWorker fetch response status:", response.status);
    console.log("[XXDK] channelsIndexedDbWorker fetch content-type:", response.headers.get('content-type'));

    if (!response.ok) {
      const text = await response.text();
      console.error("[XXDK] channelsIndexedDbWorker fetch failed. Response body:", text.substring(0, 500));
      throw new Error(`Failed to fetch WASM: ${response.status} ${response.statusText}`);
    }

    return WebAssembly.instantiateStreaming(response, go.importObject);
  }).then(async (result) => {
    go.run(result.instance);
    await isReady;
    console.info("[XXDK] channelsIndexedDbWorker started");
  }).catch((err) => {
    console.error("[XXDK] channelsIndexedDbWorker ERROR:", err);
    console.error("[XXDK] channelsIndexedDbWorker stack:", err.stack);
  });
}

export function startDmIndexedDbWorker(wasm) {
  // Prevent duplicate worker initialization
  if (self.__xxdkDmWorkerStarted) {
    console.warn("[XXDK] dmIndexedDbWorker already started, ignoring duplicate call");
    return;
  }
  self.__xxdkDmWorkerStarted = true;

  console.trace("[XXDK] dmIndexedDbWorker loading from: " + wasm.toString());
  const isReady = new Promise((resolve) => {
    self.onWasmInitialized = resolve;
  });
  const go = new Go();
  go.argv = [
    '--workerType=dm',
    '--logLevel=2',
    '--threadLogLevel=2',
  ]

  fetch(wasm).then(async (response) => {
    console.log("[XXDK] dmIndexedDbWorker fetch response status:", response.status);
    console.log("[XXDK] dmIndexedDbWorker fetch content-type:", response.headers.get('content-type'));

    if (!response.ok) {
      const text = await response.text();
      console.error("[XXDK] dmIndexedDbWorker fetch failed. Response body:", text.substring(0, 500));
      throw new Error(`Failed to fetch WASM: ${response.status} ${response.statusText}`);
    }

    return WebAssembly.instantiateStreaming(response, go.importObject);
  }).then(async (result) => {
    go.run(result.instance);
    await isReady;
    console.info("[XXDK] dmIndexedDbWorker started");
  }).catch((err) => {
    console.error("[XXDK] dmIndexedDbWorker ERROR:", err);
    console.error("[XXDK] dmIndexedDbWorker stack:", err.stack);
  });
}

/**
 * KV Worker - Pure JavaScript worker for IndexedDB key-value storage.
 * This function is serialized via toString() and loaded as a blob URL.
 * All code must be self-contained (no external imports).
 */
export function startKVWorker() {
  const KV_WORKER_VERSION = '4.0.0';
  const Ops = {
    Get: 'KVGet',
    Set: 'KVSet',
    Delete: 'KVDelete',
    Keys: 'KVKeys',
    Clear: 'KVClear',
    Init: 'KVInit',
    Ready: 'ready',
    RegisterPort: 'KVRegisterPort',
  };

  class KVWorkerService {
    constructor() {
      this.db = null;
      this.dbName = '';
      this.storeName = 'kv';
      this.ports = new Map();
      this.initialized = false;
    }

    async init(baseName) {
      if (this.initialized && this.dbName === `${baseName}_ekv`) {
        console.log(`[KVWorker] Already initialized for ${baseName}`);
        return;
      }
      this.dbName = `${baseName}_ekv`;
      console.log(`[KVWorker] v${KV_WORKER_VERSION} Initializing database: ${this.dbName}`);
      await this.migrateIfNeeded(baseName);
      this.db = await this.openDatabase();
      this.initialized = true;
      console.log(`[KVWorker] Database ${this.dbName} initialized`);
    }

    async migrateIfNeeded(baseName) {
      if (!indexedDB.databases) {
        console.log('[KVWorker] indexedDB.databases() not supported, skipping migration check');
        return;
      }
      try {
        const databases = await indexedDB.databases();
        const newDbName = `${baseName}_ekv`;
        if (databases.some(d => d.name === newDbName)) {
          console.log(`[KVWorker] Database ${newDbName} already exists, no migration needed`);
          return;
        }
        const oldDbName = baseName;
        if (databases.some(d => d.name === oldDbName)) {
          console.log(`[KVWorker] Found old database ${oldDbName}, migrating to ${newDbName}`);
          await this.performMigration(oldDbName, newDbName);
        }
      } catch (err) {
        console.error('[KVWorker] Migration check failed:', err);
      }
    }

    async performMigration(oldDbName, newDbName) {
      return new Promise((resolve) => {
        const oldRequest = indexedDB.open(oldDbName);
        oldRequest.onerror = () => {
          console.error('[KVWorker] Failed to open old database:', oldRequest.error);
          resolve();
        };
        oldRequest.onsuccess = async () => {
          const oldDb = oldRequest.result;
          if (!oldDb.objectStoreNames.contains('states')) {
            console.log('[KVWorker] Old database does not have states store, skipping migration');
            oldDb.close();
            resolve();
            return;
          }
          try {
            const entries = [];
            const readTx = oldDb.transaction(['states'], 'readonly');
            const readStore = readTx.objectStore('states');
            const cursorRequest = readStore.openCursor();
            cursorRequest.onsuccess = (event) => {
              const cursor = event.target.result;
              if (cursor) {
                const record = cursor.value;
                let value;
                if (typeof record.value === 'string') {
                  value = record.value;
                } else if (record.value instanceof Uint8Array) {
                  let binary = '';
                  for (let i = 0; i < record.value.length; i++) {
                    binary += String.fromCharCode(record.value[i]);
                  }
                  value = btoa(binary);
                } else {
                  value = String(record.value);
                }
                entries.push({ key: record.id, value });
                cursor.continue();
              } else {
                oldDb.close();
                this.writeToNewDatabase(newDbName, entries).then(resolve);
              }
            };
            cursorRequest.onerror = () => {
              console.error('[KVWorker] Failed to read from old database:', cursorRequest.error);
              oldDb.close();
              resolve();
            };
          } catch (err) {
            console.error('[KVWorker] Migration failed:', err);
            oldDb.close();
            resolve();
          }
        };
      });
    }

    async writeToNewDatabase(newDbName, entries) {
      if (entries.length === 0) {
        console.log('[KVWorker] No entries to migrate');
        return;
      }
      return new Promise((resolve, reject) => {
        const request = indexedDB.open(newDbName, 1);
        request.onerror = () => reject(request.error);
        request.onupgradeneeded = (event) => {
          const database = event.target.result;
          if (!database.objectStoreNames.contains(this.storeName)) {
            database.createObjectStore(this.storeName);
          }
        };
        request.onsuccess = () => {
          const newDb = request.result;
          const writeTx = newDb.transaction([this.storeName], 'readwrite');
          const writeStore = writeTx.objectStore(this.storeName);
          let written = 0;
          for (const entry of entries) {
            const putRequest = writeStore.put(entry.value, entry.key);
            putRequest.onsuccess = () => {
              written++;
              if (written === entries.length) {
                console.log(`[KVWorker] Migration complete: ${written} entries migrated`);
              }
            };
          }
          writeTx.oncomplete = () => {
            newDb.close();
            resolve();
          };
          writeTx.onerror = () => {
            console.error('[KVWorker] Migration write failed:', writeTx.error);
            newDb.close();
            reject(writeTx.error);
          };
        };
      });
    }

    openDatabase() {
      return new Promise((resolve, reject) => {
        const request = indexedDB.open(this.dbName, 1);
        request.onerror = () => reject(request.error);
        request.onsuccess = () => {
          console.log(`[KVWorker] Database ${this.dbName} opened successfully`);
          resolve(request.result);
        };
        request.onupgradeneeded = (event) => {
          const database = event.target.result;
          if (!database.objectStoreNames.contains(this.storeName)) {
            console.log(`[KVWorker] Creating object store: ${this.storeName}`);
            database.createObjectStore(this.storeName);
          }
        };
      });
    }

    async get(key) {
      if (!this.db) throw new Error('Database not initialized');
      return new Promise((resolve, reject) => {
        const transaction = this.db.transaction([this.storeName], 'readonly');
        const store = transaction.objectStore(this.storeName);
        const request = store.get(key);
        request.onerror = () => reject(request.error);
        request.onsuccess = () => {
          const value = request.result;
          if (value === undefined) {
            reject(new Error(`key does not exist: ${key}`));
          } else {
            resolve(value);
          }
        };
      });
    }

    async set(key, value) {
      if (!this.db) throw new Error('Database not initialized');
      return new Promise((resolve, reject) => {
        const transaction = this.db.transaction([this.storeName], 'readwrite');
        const store = transaction.objectStore(this.storeName);
        const request = store.put(value, key);
        request.onerror = () => reject(request.error);
        request.onsuccess = () => resolve();
      });
    }

    async delete(key) {
      if (!this.db) throw new Error('Database not initialized');
      return new Promise((resolve, reject) => {
        const transaction = this.db.transaction([this.storeName], 'readwrite');
        const store = transaction.objectStore(this.storeName);
        const request = store.delete(key);
        request.onerror = () => reject(request.error);
        request.onsuccess = () => resolve();
      });
    }

    async keys() {
      if (!this.db) throw new Error('Database not initialized');
      return new Promise((resolve, reject) => {
        const transaction = this.db.transaction([this.storeName], 'readonly');
        const store = transaction.objectStore(this.storeName);
        const request = store.getAllKeys();
        request.onerror = () => reject(request.error);
        request.onsuccess = () => resolve(request.result);
      });
    }

    async clear() {
      if (!this.db) throw new Error('Database not initialized');
      return new Promise((resolve, reject) => {
        const transaction = this.db.transaction([this.storeName], 'readwrite');
        const store = transaction.objectStore(this.storeName);
        const request = store.clear();
        request.onerror = () => reject(request.error);
        request.onsuccess = () => resolve();
      });
    }

    close() {
      if (this.db) {
        console.log(`[KVWorker] Closing database: ${this.dbName}`);
        this.db.close();
        this.db = null;
        this.initialized = false;
      }
      this.ports.forEach((port, channel) => {
        console.log(`[KVWorker] Closing port for channel: ${channel}`);
        port.close();
      });
      this.ports.clear();
    }

    registerPort(channel, port) {
      console.log(`[KVWorker] Registering port for channel: ${channel}`);
      this.ports.set(channel, port);
      port.onmessage = async (event) => {
        await this.handleMessage(event.data, (response) => {
          port.postMessage(response);
        });
      };
      port.start();
    }

    objectToBytes(obj) {
      const keys = Object.keys(obj).filter(k => !isNaN(Number(k))).sort((a, b) => Number(a) - Number(b));
      const bytes = new Uint8Array(keys.length);
      for (let i = 0; i < keys.length; i++) {
        bytes[i] = obj[keys[i]];
      }
      return bytes;
    }

    decodeLegacyPayload(data) {
      if (!data) return {};
      try {
        const parsed = JSON.parse(data);
        if (typeof parsed === 'object' && parsed !== null) return parsed;
      } catch {}
      try {
        const decoded = atob(data);
        return JSON.parse(decoded);
      } catch {}
      return {};
    }

    async handleMessage(data, sendResponse) {
      try {
        let text;
        if (typeof data === 'string') {
          text = data;
        } else if (data instanceof ArrayBuffer) {
          text = new TextDecoder().decode(new Uint8Array(data));
        } else if (data instanceof Uint8Array) {
          text = new TextDecoder().decode(data);
        } else if (ArrayBuffer.isView(data)) {
          text = new TextDecoder().decode(new Uint8Array(data.buffer, data.byteOffset, data.byteLength));
        } else if (typeof data === 'object' && data !== null && '0' in data) {
          const bytes = this.objectToBytes(data);
          text = new TextDecoder().decode(bytes);
        } else {
          console.error('[KVWorker] Unexpected message type:', typeof data);
          return;
        }
        const message = JSON.parse(text);
        await this.processMessage(message, sendResponse);
      } catch (err) {
        console.error('[KVWorker] Failed to handle message:', err);
      }
    }

    async processMessage(message, sendResponse) {
      if (message.response) return;
      const response = { op: message.op, id: message.id, response: true };
      try {
        const hasDirectFields = message.key !== undefined || message.dbName !== undefined || message.value !== undefined;
        let key, value, dbName;
        if (hasDirectFields) {
          key = message.key;
          value = message.value;
          dbName = message.dbName;
        } else if (message.data) {
          const payload = this.decodeLegacyPayload(message.data);
          key = payload.key;
          value = payload.value;
          dbName = payload.dbName;
        }
        switch (message.op) {
          case Ops.Init:
            if (!dbName) throw new Error('dbName is required for init');
            await this.init(dbName);
            break;
          case Ops.Get:
            if (!key) throw new Error('key is required for get');
            response.value = await this.get(key);
            break;
          case Ops.Set:
            if (!key) throw new Error('key is required for set');
            if (value === undefined) throw new Error('value is required for set');
            await this.set(key, value);
            break;
          case Ops.Delete:
            if (!key) throw new Error('key is required for delete');
            await this.delete(key);
            break;
          case Ops.Keys:
            const keysList = await this.keys();
            response.value = JSON.stringify(keysList);
            break;
          case Ops.Clear:
            await this.clear();
            break;
          default:
            response.error = `Unknown op: ${message.op}`;
        }
      } catch (err) {
        response.error = err instanceof Error ? err.message : String(err);
      }
      sendResponse(JSON.stringify(response));
    }
  }

  const kvService = new KVWorkerService();

  self.onmessage = async (event) => {
    const data = event.data;
    if (event.ports && event.ports.length > 0) {
      try {
        let text;
        if (typeof data === 'string') {
          text = data;
        } else if (data instanceof ArrayBuffer || data instanceof Uint8Array) {
          text = new TextDecoder().decode(data instanceof ArrayBuffer ? new Uint8Array(data) : data);
        } else {
          console.error('[KVWorker] Unexpected port registration data type');
          return;
        }
        const message = JSON.parse(text);
        if ('channel' in message) {
          kvService.registerPort(message.channel, event.ports[0]);
          return;
        }
      } catch (err) {
        console.error('[KVWorker] Failed to register port:', err);
      }
      return;
    }
    const sendResponse = (response) => {
      self.postMessage(response);
    };
    await kvService.handleMessage(data, sendResponse);
  };

  self.addEventListener('close', () => {
    console.log(`[KVWorker] v${KV_WORKER_VERSION} - Worker closing`);
    kvService.close();
  });

  console.log(`[KVWorker] v${KV_WORKER_VERSION} - Worker initialized (string-only storage)`);
  self.postMessage(JSON.stringify({ op: Ops.Ready, id: 0, response: false, data: '' }));
}
