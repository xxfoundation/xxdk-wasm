/* eslint-disable no-console */
import { encoder } from '../utils/index';

export enum AccountSyncStatus {
  NotSynced = 'NotSynced',
  Synced = 'Synced',
  Ignore = 'Ignored'
}

export enum AccountSyncService {
  None = 'None',
  Google = 'Google',
  Dropbox = 'Dropbox'
}

export enum OperationType {
  Created = 0,
  Updated = 1,
  Deleted = 2 
}

export type KVEntry = {
  version: number;
  timestamp: string;
  data: string;
}

type KeyChangedByRemoteCallback = {
  Callback: (
    key: string,
    oldEntry: Uint8Array,
    newEntry: Uint8Array,
    operationType: OperationType
  ) => void;
}

export const KV_VERSION = 0;

export type RemoteKV = {
  Get: (key: string, version: number) => Promise<Uint8Array>;
  Delete: (key: string, version: number) => Promise<void>;
  Set: (key: string, encodedKVMapEntry: Uint8Array) => Promise<void>;
  ListenOnRemoteKey: (key: string, version: number, onChange: KeyChangedByRemoteCallback) => number;
  DeleteRemoteKeyListener: (key: string, id: number) => void;
}

export interface RemoteStoreServiceWrapper {
  service: AccountSyncService;
  Read: (path: string) => Promise<Uint8Array | null>;
  Write: (path: string, data: Uint8Array) => Promise<void>;
  GetLastModified: (path: string) => Promise<string | null>;
  ReadDir: (path: string) => Promise<string[]>;
  DeleteAll: () => Promise<void>; 
}

export class RemoteStore {
  store: RemoteStoreServiceWrapper;

  lastWrite: string | null = null;

  service: AccountSyncService;

  constructor(store: RemoteStoreServiceWrapper) {
    this.service = store.service;
    this.store = store;
  }

  async Read(path: string) {
    const read = await this.store.Read(path);
    console.log(`[KV] Read path ${path}`, read);
    return read;
  }

  Write(path: string, data: Uint8Array) {
    this.lastWrite = new Date().toISOString();
    console.log('[KV] Write path', path);
    return this.store.Write(path, data);
  }

  GetLastWrite() {
    console.log('[KV] GetLastWrite', this.lastWrite);
    return this.lastWrite;
  }

  async GetLastModified(path: string) {
    const date = await this.store.GetLastModified(path);
    console.log('[KV] GetLastModified path', path, 'date', date);
    return date && new Date(date).toISOString();
  }

  async ReadDir(path: string) {
    console.log(`[KV] ReadDir path ${path}`);
    const dirs = await this.store.ReadDir(path);
    console.log(`[KV] ReadDir dirs ${dirs.join(', ')}`);
    return encoder.encode(JSON.stringify(dirs));
  }

  DeleteAll() {
    console.log('[KV] DeleteAll');
    return this.store.DeleteAll();
  }
}
