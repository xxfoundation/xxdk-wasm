/**
 * Generic Key-Value storage for xxDK
 *
 * Provides KV storage with a string-only interface.
 * Callers are responsible for encoding/decoding values (e.g., base32768 for bytes).
 *
 * - WorkerKV: Uses a dedicated Web Worker for IndexedDB (recommended)
 * - InMemoryKV: In-memory storage for testing
 */

import { kvWorkerPath } from './paths';

// Common interface for KV storage - string only
export interface KVStore {
  Get(key: string): Promise<string>;
  Set(key: string, value: string): Promise<void>;
  Delete(key: string): Promise<void>;
  Keys(): Promise<string[]>;
  Clear(): Promise<void>;
}

// Message structure with direct fields (no nested encoding)
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
} as const;

/**
 * WorkerKV implements KVStore by communicating with a KV Worker.
 * This keeps IndexedDB operations off the main thread.
 * All values are strings - encoding is done by the caller.
 */
class WorkerKV implements KVStore {
  private _worker: Worker;
  private dbName: string;
  private messageId = 0;
  private pendingRequests = new Map<number, {
    resolve: (value: string) => void;
    reject: (error: Error) => void;
  }>();
  private initialized = false;
  private initPromise: Promise<void> | null = null;

  constructor(workerUrl: string, storageDir: string) {
    this.dbName = storageDir;
    console.log(`[WorkerKV] Creating worker for ${storageDir} from ${workerUrl}`);

    this._worker = new Worker(workerUrl, { name: 'kvWorker' });

    // Setup message handler
    this._worker.onmessage = (event: MessageEvent) => {
      this.handleMessage(event.data);
    };

    this._worker.onerror = (error: ErrorEvent) => {
      console.error('[WorkerKV] Worker error:', error);
    };
  }

  /**
   * Get the underlying Worker object for use with MessageChannel.
   */
  get worker(): Worker {
    return this._worker;
  }

  /**
   * Initialize the worker with the database name.
   * Must be called before any other operations.
   */
  async init(): Promise<void> {
    if (this.initialized) return;
    if (this.initPromise) return this.initPromise;

    this.initPromise = new Promise<void>((resolve, reject) => {
      const id = this.getNextId();

      this.pendingRequests.set(id, {
        resolve: () => {
          this.initialized = true;
          resolve();
        },
        reject
      });

      // Use direct dbName field
      const message: KVMessage = {
        op: Ops.Init,
        id,
        response: false,
        dbName: this.dbName
      };
      this._worker.postMessage(JSON.stringify(message));
    });

    return this.initPromise;
  }

  private async ensureInitialized(): Promise<void> {
    if (!this.initialized) {
      await this.init();
    }
  }

  async Get(key: string): Promise<string> {
    await this.ensureInitialized();

    return new Promise((resolve, reject) => {
      const id = this.getNextId();
      this.pendingRequests.set(id, { resolve, reject });
      // Use direct key field
      const message: KVMessage = {
        op: Ops.Get,
        id,
        response: false,
        key
      };
      this._worker.postMessage(JSON.stringify(message));
    });
  }

  async Set(key: string, value: string): Promise<void> {
    await this.ensureInitialized();

    return new Promise((resolve, reject) => {
      const id = this.getNextId();
      this.pendingRequests.set(id, {
        resolve: () => resolve(),
        reject
      });
      // Use direct key and value fields
      const message: KVMessage = {
        op: Ops.Set,
        id,
        response: false,
        key,
        value
      };
      this._worker.postMessage(JSON.stringify(message));
    });
  }

  async Delete(key: string): Promise<void> {
    await this.ensureInitialized();

    return new Promise((resolve, reject) => {
      const id = this.getNextId();
      this.pendingRequests.set(id, {
        resolve: () => resolve(),
        reject
      });
      // Use direct key field
      const message: KVMessage = {
        op: Ops.Delete,
        id,
        response: false,
        key
      };
      this._worker.postMessage(JSON.stringify(message));
    });
  }

  async Keys(): Promise<string[]> {
    await this.ensureInitialized();

    return new Promise((resolve, reject) => {
      const id = this.getNextId();
      this.pendingRequests.set(id, {
        resolve: (data: string) => {
          try {
            resolve(JSON.parse(data) as string[]);
          } catch {
            resolve([]);
          }
        },
        reject
      });
      const message: KVMessage = {
        op: Ops.Keys,
        id,
        response: false
      };
      this._worker.postMessage(JSON.stringify(message));
    });
  }

  async Clear(): Promise<void> {
    await this.ensureInitialized();

    return new Promise((resolve, reject) => {
      const id = this.getNextId();
      this.pendingRequests.set(id, {
        resolve: () => resolve(),
        reject
      });
      const message: KVMessage = {
        op: Ops.Clear,
        id,
        response: false
      };
      this._worker.postMessage(JSON.stringify(message));
    });
  }

  private handleMessage(data: unknown): void {
    try {
      // Messages are always JSON strings
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
        // Handle structured clone artifact: Uint8Array serialized as {0: 65, 1: 66, ...}
        const bytes = this.objectToBytes(data as Record<string, number>);
        text = new TextDecoder().decode(bytes);
      } else {
        console.error('[WorkerKV] Unexpected message data type:', typeof data, data);
        return;
      }

      const message: KVMessage = JSON.parse(text);

      // Only handle responses
      if (!message.response) {
        if (message.op === Ops.Ready) {
          console.log('[WorkerKV] Worker ready');
        }
        return;
      }

      const pending = this.pendingRequests.get(message.id);
      if (!pending) {
        console.warn('[WorkerKV] No pending request for id:', message.id);
        return;
      }

      this.pendingRequests.delete(message.id);

      // Check for error in response (direct field)
      if (message.error) {
        pending.reject(new Error(message.error));
        return;
      }

      // Get value from direct field
      if (message.value !== undefined) {
        pending.resolve(message.value);
        return;
      }

      // Empty response means success for void operations
      pending.resolve('');
    } catch (err) {
      console.error('[WorkerKV] Failed to handle message:', err);
    }
  }

  private getNextId(): number {
    return this.messageId++;
  }

  /**
   * Convert object with numeric keys (structured clone artifact) to Uint8Array.
   */
  private objectToBytes(obj: Record<string, number>): Uint8Array {
    const keys = Object.keys(obj).filter(k => !isNaN(Number(k))).sort((a, b) => Number(a) - Number(b));
    const bytes = new Uint8Array(keys.length);
    for (let i = 0; i < keys.length; i++) {
      bytes[i] = obj[keys[i]];
    }
    return bytes;
  }

  /**
   * Terminate the worker
   */
  terminate(): void {
    this._worker.terminate();
  }
}

// In-memory Map implementation (synchronous, no I/O overhead)
class InMemoryKV implements KVStore {
  private store: Map<string, string>;
  private prefix: string;

  constructor(storageDir: string) {
    this.prefix = `xxdk_${storageDir}_`;
    this.store = new Map();
    console.log('[InMemoryKV] Created for:', storageDir);
  }

  async Get(key: string): Promise<string> {
    const fullKey = this.prefix + key;
    const value = this.store.get(fullKey);

    if (value === undefined) {
      throw new Error(`key does not exist: ${key}`);
    }

    return value;
  }

  async Set(key: string, value: string): Promise<void> {
    const fullKey = this.prefix + key;
    this.store.set(fullKey, value);
  }

  async Delete(key: string): Promise<void> {
    const fullKey = this.prefix + key;
    this.store.delete(fullKey);
  }

  async Keys(): Promise<string[]> {
    const keys: string[] = [];

    this.store.forEach((_, key) => {
      if (key.startsWith(this.prefix)) {
        keys.push(key.substring(this.prefix.length));
      }
    });

    return keys;
  }

  async Clear(): Promise<void> {
    const keysToDelete: string[] = [];
    this.store.forEach((_, key) => {
      if (key.startsWith(this.prefix)) {
        keysToDelete.push(key);
      }
    });
    keysToDelete.forEach(key => this.store.delete(key));
  }
}

// Store manager for tracking KV stores per storageDir
interface StoreManager {
  kvInstance: KVStore;
  initPromise?: Promise<KVStore>;
}

// Module-level state
const stores = new Map<string, StoreManager>();
let kvWorkerUrl: string | null = null;
let globalWorkerKV: WorkerKV | null = null;

// Global key for storing worker reference (shared between TS and Go)
const KV_WORKER_GLOBAL_KEY = '__xxdkKVWorker';

/**
 * Set the KV worker URL. Must be called before createKVStore.
 */
export function setKVWorkerUrl(url: string): void {
  kvWorkerUrl = url;
  console.log('[KV] Worker URL set to:', url);
}

/**
 * Get the underlying Worker object for Go WASM to use.
 * Returns null if not initialized yet.
 */
export function getKVWorker(): Worker | null {
  return globalWorkerKV?.worker ?? null;
}

/**
 * Get or create KV store for storageDir.
 * Lazy-loads the KV Worker URL if not already set.
 * Stores the worker on globalThis for Go WASM to access.
 */
async function getKVStore(storageDir: string): Promise<KVStore> {
  // Lazy-load the worker URL if not set
  if (!kvWorkerUrl) {
    console.log('[KV] Lazy-loading KV Worker URL...');
    kvWorkerUrl = await kvWorkerPath();
    console.log('[KV] KV Worker URL loaded:', kvWorkerUrl);
  }

  // Check if already initialized
  const existing = stores.get(storageDir);
  if (existing?.kvInstance) {
    return existing.kvInstance;
  }

  // Check if initialization in progress
  if (existing?.initPromise) {
    console.log(`[KV] KV store initialization in progress for: ${storageDir}, waiting...`);
    return await existing.initPromise;
  }

  // Start new initialization
  console.log(`[KV] Creating KV store for: ${storageDir}`);
  const initPromise = (async () => {
    const workerKv = new WorkerKV(kvWorkerUrl!, storageDir);
    await workerKv.init();

    // Store global reference on globalThis for Go WASM to access
    if (!globalWorkerKV) {
      globalWorkerKV = workerKv;
      // Store on globalThis so Go can find it
      (globalThis as any)[KV_WORKER_GLOBAL_KEY] = workerKv.worker;
      console.log(`[KV] Worker stored on globalThis.${KV_WORKER_GLOBAL_KEY}`);
    }

    const manager: StoreManager = {
      kvInstance: workerKv,
      initPromise: undefined
    };
    stores.set(storageDir, manager);
    return workerKv;
  })();

  // Store promise to prevent race conditions
  stores.set(storageDir, {
    kvInstance: null as any,
    initPromise
  });

  return await initPromise;
}

// Factory function to create KV instance
export async function createKVStore(storageDir: string): Promise<KVStore> {
  return await getKVStore(storageDir);
}

/**
 * Cleanup all KV workers and connections.
 * Call this before deleting IndexedDB databases.
 */
export function cleanupKVWorkers(): void {
  console.log('[KV] Cleaning up all KV workers...');

  // Terminate global worker
  if (globalWorkerKV) {
    globalWorkerKV.terminate();
    globalWorkerKV = null;
    console.log('[KV] Global WorkerKV terminated');
  }

  // Clear stores map
  stores.forEach((manager, key) => {
    if (manager.kvInstance && 'terminate' in manager.kvInstance) {
      (manager.kvInstance as WorkerKV).terminate();
      console.log(`[KV] Terminated worker for: ${key}`);
    }
  });
  stores.clear();

  // Reset URL so next createKVStore will reinitialize
  kvWorkerUrl = null;

  console.log('[KV] All KV workers cleaned up');
}

// Global registry to reuse InMemoryKV instances per storageDir
const inMemoryKVInstances = new Map<string, InMemoryKV>();

// Factory function to create in-memory KV instance (for testing)
export function createInMemoryKV(storageDir: string): KVStore {
  let instance = inMemoryKVInstances.get(storageDir);
  if (!instance) {
    instance = new InMemoryKV(storageDir);
    inMemoryKVInstances.set(storageDir, instance);
  }
  return instance;
}
