import { RemoteKVWrapper } from '@contexts/remote-kv-context';
import { CMix, Message, MessagePinEvent, MessageUnPinEvent, RemoteStore, TypedEventEmitter } from '@types';
import { makeEventAwaiter, makeEventHook, makeListenerHook } from '@utils/index';
import { AccountSyncService } from 'src/hooks/useAccountSync';
import EventEmitter from 'events';
import { ChannelManager } from '@contexts/network-client-context';

export enum AppEvents {
  MESSAGE_PINNED = 'pinned',
  MESSAGE_UNPINNED = 'unpinned',
  GOOGLE_TOKEN = 'google-token',
  DROPBOX_TOKEN = 'dropbox-token',
  CMIX_LOADED = 'cmix-loaded',
  CHANNEL_MANAGER_LOADED = 'channel-manager-loaded',
  REMOTE_STORE_INITIALIZED = 'remote-store-initialized',
  CMIX_SYNCED = 'cmix-synced',
  PASSWORD_ENTERED = 'password-entered',
  PASSWORD_DECRYPTED = 'password-decrypted',
  REMOTE_KV_INITIALIZED = 'remote-kv-initialized',
  DM_NOTIFICATION_UPDATE = 'dm-notifications-update',
  MESSAGE_PROCESSED = 'message-processed',
  DM_PROCESSED = 'dm-processed',
  NEW_SYNC_CMIX_FAILED = 'new-sync-cmix-failed',
  EMOJI_SELECTED = 'emoji-selected',
  MESSAGES_FETCHED = 'messages-fetched'
}

export type AppEventHandlers = {
  [AppEvents.MESSAGE_PINNED]: (event: MessagePinEvent) => void;
  [AppEvents.MESSAGE_UNPINNED]: (event: MessageUnPinEvent) => void;
  [AppEvents.GOOGLE_TOKEN]: (event: string) => void;
  [AppEvents.DROPBOX_TOKEN]: (event: string) => void;
  [AppEvents.REMOTE_STORE_INITIALIZED]: (remoteStore: RemoteStore) => void;
  [AppEvents.CMIX_LOADED]: (cmix: CMix) => void;
  [AppEvents.PASSWORD_ENTERED]: (rawPassword: string) => void;
  [AppEvents.PASSWORD_DECRYPTED]: (decrypted: Uint8Array, rawPassword: string) => void;
  [AppEvents.CMIX_SYNCED]: (service: AccountSyncService) => void;
  [AppEvents.REMOTE_KV_INITIALIZED]: (kv: RemoteKVWrapper) => void;
  [AppEvents.CHANNEL_MANAGER_LOADED]: (channelManager: ChannelManager) => void;
  [AppEvents.MESSAGE_PROCESSED]: (message: Message, oldMessage?: Message) => void;
  [AppEvents.NEW_SYNC_CMIX_FAILED]: () => void;
  [AppEvents.DM_PROCESSED]: (message: Message) => void;
  [AppEvents.EMOJI_SELECTED]: (emoji: string) => void;
  [AppEvents.MESSAGES_FETCHED]: (fetched: boolean) => void;
}

export const appBus = new EventEmitter() as TypedEventEmitter<AppEventHandlers>;

export const useAppEventListener = makeListenerHook(appBus);

export const awaitAppEvent = makeEventAwaiter(appBus);

export const useAppEventValue = makeEventHook(appBus);

