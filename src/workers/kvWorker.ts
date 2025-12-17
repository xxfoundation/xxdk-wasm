////////////////////////////////////////////////////////////////////////////////
// Copyright © 2024 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

// Version for cache busting and debugging
const KV_WORKER_VERSION = '4.0.0';

/**
 * KV Worker - Centralized IndexedDB storage service
 *
 * This pure JavaScript worker handles all KV storage operations.
 * All values are strings - callers are responsible for encoding/decoding
 * (e.g., base32768 for bytes).
 *
 * Accepts connections via MessageChannel from:
 * - Main thread (Go WASM)
 * - Channels worker
 * - DM worker
 *
 * Message format (direct fields, no nested encoding):
 * {
 *   op: string,        // Operation (e.g., "KVGet", "KVSet")
 *   id: number,        // Unique message ID for response correlation
 *   response: boolean, // true if this is a response
 *   key?: string,      // Key for get/set/delete
 *   value?: string,    // Value for set or get response (base32768-encoded)
 *   dbName?: string,   // Database name for init
 *   error?: string     // Error message if operation failed
 * }
 */

// Message structure with direct fields
interface KVMessage {
  op: string;
  id: number;
  response: boolean;
  // Request fields
  key?: string;
  value?: string;
  dbName?: string;
  // Response fields
  error?: string;
  // Legacy data field for backwards compatibility
  data?: string;
}

// KV operations
const Ops = {
  Get: 'KVGet',
  Set: 'KVSet',
  Delete: 'KVDelete',
  Keys: 'KVKeys',
  Clear: 'KVClear',
  Init: 'KVInit',
  Ready: 'ready',
  RegisterPort: 'KVRegisterPort',
} as const;

/**
 * KVWorkerService handles all IndexedDB operations with string values.
 */
class KVWorkerService {
  private db: IDBDatabase | null = null;
  private dbName: string = '';
  private storeName = 'kv';
  private ports: Map<string, MessagePort> = new Map();
  private initialized = false;

  /**
   * Initialize the KV service with a database name
   */
  async init(baseName: string): Promise<void> {
    if (this.initialized && this.dbName === `${baseName}_ekv`) {
      console.log(`[KVWorker] Already initialized for ${baseName}`);
      return;
    }

    this.dbName = `${baseName}_ekv`;
    console.log(`[KVWorker] v${KV_WORKER_VERSION} Initializing database: ${this.dbName}`);

    // Check for migration from old database
    await this.migrateIfNeeded(baseName);

    // Open database
    this.db = await this.openDatabase();
    this.initialized = true;
    console.log(`[KVWorker] Database ${this.dbName} initialized`);
  }

  /**
   * Check if migration is needed and perform it
   */
  private async migrateIfNeeded(baseName: string): Promise<void> {
    if (!indexedDB.databases) {
      console.log('[KVWorker] indexedDB.databases() not supported, skipping migration check');
      return;
    }

    try {
      const databases = await indexedDB.databases();
      const newDbName = `${baseName}_ekv`;

      const newExists = databases.some(d => d.name === newDbName);
      if (newExists) {
        console.log(`[KVWorker] Database ${newDbName} already exists, no migration needed`);
        return;
      }

      // Check for old database (without _ekv suffix)
      const oldDbName = baseName;
      const oldExists = databases.some(d => d.name === oldDbName);

      if (oldExists) {
        console.log(`[KVWorker] Found old database ${oldDbName}, migrating to ${newDbName}`);
        await this.performMigration(oldDbName, newDbName);
      }
    } catch (err) {
      console.error('[KVWorker] Migration check failed:', err);
    }
  }

  /**
   * Perform migration from old database format to new format
   */
  private async performMigration(oldDbName: string, newDbName: string): Promise<void> {
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
          const entries: Array<{key: string, value: string}> = [];
          const readTx = oldDb.transaction(['states'], 'readonly');
          const readStore = readTx.objectStore('states');
          const cursorRequest = readStore.openCursor();

          cursorRequest.onsuccess = (event) => {
            const cursor = (event.target as IDBRequest<IDBCursorWithValue>).result;
            if (cursor) {
              const record = cursor.value;
              // Old format stored values - convert to string if needed
              let value: string;
              if (typeof record.value === 'string') {
                value = record.value;
              } else if (record.value instanceof Uint8Array) {
                // Convert old Uint8Array to base64 string for migration
                value = this.uint8ArrayToBase64(record.value);
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

  private uint8ArrayToBase64(bytes: Uint8Array): string {
    let binary = '';
    for (let i = 0; i < bytes.length; i++) {
      binary += String.fromCharCode(bytes[i]);
    }
    return btoa(binary);
  }

  /**
   * Write migrated entries to new database
   */
  private async writeToNewDatabase(
    newDbName: string,
    entries: Array<{key: string, value: string}>
  ): Promise<void> {
    if (entries.length === 0) {
      console.log('[KVWorker] No entries to migrate');
      return;
    }

    return new Promise((resolve, reject) => {
      // Version 2: ensures 'kv' store exists
      const request = indexedDB.open(newDbName, 2);

      request.onerror = () => reject(request.error);

      request.onupgradeneeded = (event) => {
        const database = (event.target as IDBOpenDBRequest).result;
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

  /**
   * Open the database
   */
  private openDatabase(): Promise<IDBDatabase> {
    return new Promise((resolve, reject) => {
      // Version 2: ensures 'kv' store exists (fixes databases created without it)
      const request = indexedDB.open(this.dbName, 2);

      request.onerror = () => reject(request.error);

      request.onsuccess = () => {
        console.log(`[KVWorker] Database ${this.dbName} opened successfully`);
        resolve(request.result);
      };

      request.onupgradeneeded = (event) => {
        const database = (event.target as IDBOpenDBRequest).result;
        if (!database.objectStoreNames.contains(this.storeName)) {
          console.log(`[KVWorker] Creating object store: ${this.storeName}`);
          database.createObjectStore(this.storeName);
        }
      };
    });
  }

  /**
   * Get a string value by key
   */
  async get(key: string): Promise<string> {
    if (!this.db) {
      throw new Error('Database not initialized');
    }

    return new Promise((resolve, reject) => {
      const transaction = this.db!.transaction([this.storeName], 'readonly');
      const store = transaction.objectStore(this.storeName);
      const request = store.get(key);

      request.onerror = () => reject(request.error);

      request.onsuccess = () => {
        const value = request.result;
        if (value === undefined) {
          reject(new Error(`key does not exist: ${key}`));
        } else {
          resolve(value as string);
        }
      };
    });
  }

  /**
   * Set a string value for a key
   */
  async set(key: string, value: string): Promise<void> {
    if (!this.db) {
      throw new Error('Database not initialized');
    }

    return new Promise((resolve, reject) => {
      const transaction = this.db!.transaction([this.storeName], 'readwrite');
      const store = transaction.objectStore(this.storeName);
      const request = store.put(value, key);

      request.onerror = () => reject(request.error);
      request.onsuccess = () => resolve();
    });
  }

  /**
   * Delete a key
   */
  async delete(key: string): Promise<void> {
    if (!this.db) {
      throw new Error('Database not initialized');
    }

    return new Promise((resolve, reject) => {
      const transaction = this.db!.transaction([this.storeName], 'readwrite');
      const store = transaction.objectStore(this.storeName);
      const request = store.delete(key);

      request.onerror = () => reject(request.error);
      request.onsuccess = () => resolve();
    });
  }

  /**
   * Get all keys
   */
  async keys(): Promise<string[]> {
    if (!this.db) {
      throw new Error('Database not initialized');
    }

    return new Promise((resolve, reject) => {
      const transaction = this.db!.transaction([this.storeName], 'readonly');
      const store = transaction.objectStore(this.storeName);
      const request = store.getAllKeys();

      request.onerror = () => reject(request.error);
      request.onsuccess = () => resolve(request.result as string[]);
    });
  }

  /**
   * Clear all keys from the store
   */
  async clear(): Promise<void> {
    if (!this.db) {
      throw new Error('Database not initialized');
    }

    return new Promise((resolve, reject) => {
      const transaction = this.db!.transaction([this.storeName], 'readwrite');
      const store = transaction.objectStore(this.storeName);
      const request = store.clear();

      request.onerror = () => reject(request.error);
      request.onsuccess = () => resolve();
    });
  }

  /**
   * Close the database connection.
   */
  close(): void {
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

  /**
   * Register a MessagePort from another worker (Go WASM)
   */
  registerPort(channel: string, port: MessagePort): void {
    console.log(`[KVWorker] Registering port for channel: ${channel}`);
    this.ports.set(channel, port);

    port.onmessage = async (event: MessageEvent) => {
      await this.handleMessage(event.data, (response: string) => {
        port.postMessage(response);
      });
    };

    port.start();
  }

  /**
   * Handle incoming message (string or bytes)
   */
  async handleMessage(
    data: unknown,
    sendResponse: (response: string) => void
  ): Promise<void> {
    try {
      // Parse message - handle both string and bytes
      let text: string;
      if (typeof data === 'string') {
        text = data;
      } else if (data instanceof ArrayBuffer) {
        text = new TextDecoder().decode(new Uint8Array(data));
      } else if (data instanceof Uint8Array) {
        text = new TextDecoder().decode(data);
      } else if (ArrayBuffer.isView(data)) {
        const view = data as ArrayBufferView;
        text = new TextDecoder().decode(new Uint8Array(view.buffer, view.byteOffset, view.byteLength));
      } else if (typeof data === 'object' && data !== null && '0' in (data as object)) {
        // Handle structured clone artifact
        const bytes = this.objectToBytes(data as Record<string, number>);
        text = new TextDecoder().decode(bytes);
      } else {
        console.error('[KVWorker] Unexpected message type:', typeof data);
        return;
      }

      const message: KVMessage = JSON.parse(text);
      await this.processMessage(message, sendResponse);
    } catch (err) {
      console.error('[KVWorker] Failed to handle message:', err);
    }
  }

  private objectToBytes(obj: Record<string, number>): Uint8Array {
    const keys = Object.keys(obj).filter(k => !isNaN(Number(k))).sort((a, b) => Number(a) - Number(b));
    const bytes = new Uint8Array(keys.length);
    for (let i = 0; i < keys.length; i++) {
      bytes[i] = obj[keys[i]];
    }
    return bytes;
  }

  /**
   * Decode payload from legacy message.data field.
   * Handles both plain JSON (from TypeScript) and base64-encoded JSON (from old Go).
   * Only used for backwards compatibility with old message format.
   */
  private decodeLegacyPayload(data: string): Record<string, unknown> {
    if (!data) return {};

    // Try parsing as plain JSON first (TypeScript sends this way)
    try {
      const parsed = JSON.parse(data);
      if (typeof parsed === 'object' && parsed !== null) {
        return parsed;
      }
    } catch {
      // Not valid JSON, try base64 decode
    }

    // Try base64 decode (old Go format)
    try {
      const decoded = atob(data);
      return JSON.parse(decoded);
    } catch {
      // Not valid base64 or not JSON after decode
    }

    return {};
  }

  /**
   * Process a parsed message.
   * Supports both new direct field format and legacy data field format.
   */
  private async processMessage(
    message: KVMessage,
    sendResponse: (response: string) => void
  ): Promise<void> {
    // Ignore responses
    if (message.response) {
      return;
    }

    // Build response with direct fields
    const response: KVMessage = {
      op: message.op,
      id: message.id,
      response: true
    };

    try {
      // Check if using new direct field format or legacy data format
      const hasDirectFields = message.key !== undefined ||
                             message.dbName !== undefined ||
                             message.value !== undefined;

      // Extract values - prefer direct fields, fall back to legacy data
      let key: string | undefined;
      let value: string | undefined;
      let dbName: string | undefined;

      if (hasDirectFields) {
        // New format: direct fields
        key = message.key;
        value = message.value;
        dbName = message.dbName;
      } else if (message.data) {
        // Legacy format: decode from data field
        const payload = this.decodeLegacyPayload(message.data);
        key = payload.key as string | undefined;
        value = payload.value as string | undefined;
        dbName = payload.dbName as string | undefined;
      }

      switch (message.op) {
        case Ops.Init: {
          if (!dbName) throw new Error('dbName is required for init');
          await this.init(dbName);
          break;
        }

        case Ops.Get: {
          if (!key) throw new Error('key is required for get');
          response.value = await this.get(key);
          break;
        }

        case Ops.Set: {
          if (!key) throw new Error('key is required for set');
          if (value === undefined) throw new Error('value is required for set');
          await this.set(key, value);
          break;
        }

        case Ops.Delete: {
          if (!key) throw new Error('key is required for delete');
          await this.delete(key);
          break;
        }

        case Ops.Keys: {
          const keysList = await this.keys();
          response.value = JSON.stringify(keysList);
          break;
        }

        case Ops.Clear: {
          await this.clear();
          break;
        }

        default:
          response.error = `Unknown op: ${message.op}`;
      }
    } catch (err) {
      response.error = err instanceof Error ? err.message : String(err);
    }

    sendResponse(JSON.stringify(response));
  }
}

// Create service instance
const kvService = new KVWorkerService();

// Main thread message handler
self.onmessage = async (event: MessageEvent) => {
  const data = event.data;

  // Check for MessageChannel port transfer
  if (event.ports && event.ports.length > 0) {
    try {
      let text: string;
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

  // Handle regular message
  const sendResponse = (response: string) => {
    (self as unknown as Worker).postMessage(response);
  };

  await kvService.handleMessage(data, sendResponse);
};

// Close database when worker is terminated
self.addEventListener('close', () => {
  console.log(`[KVWorker] v${KV_WORKER_VERSION} - Worker closing`);
  kvService.close();
});

self.addEventListener('beforeunload', () => {
  console.log(`[KVWorker] v${KV_WORKER_VERSION} - beforeunload`);
  kvService.close();
});

// Signal ready
console.log(`[KVWorker] v${KV_WORKER_VERSION} - Worker initialized (string-only storage)`);
(self as unknown as Worker).postMessage(JSON.stringify({
  op: Ops.Ready,
  id: 0,
  response: false,
  data: ''
}));
