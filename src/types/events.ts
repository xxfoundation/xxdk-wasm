/* eslint-disable @typescript-eslint/no-explicit-any */
import type { Message } from './messages';
import type { MessageType } from './db';
import { PrivacyLevel } from '../types';

export type Channel = {
  name: string;
  id: string;
  description: string;
  isAdmin: boolean;
  privacyLevel: PrivacyLevel | null;
  prettyPrint?: string;
}

export type ChannelId = Channel['id'];



export type MessageReceivedEvent = {
  uuid: number;
  channelId: string;
  update: boolean;
}

export type MessagePinEvent = Message;

export type MessageUnPinEvent = Message;

export type MessageDeletedEvent = {
  messageId: string;
}

export type UserMutedEvent = {
  channelId: string;
  pubkey: string;
  unmute: boolean;
}

export type DMReceivedEvent = {
  uuid: number;
  pubkey: string;
  update: boolean;
  conversationUpdated: boolean;
}

export type NicknameUpdatedEvent = {
  channelId: string;
  nickname: string;
  exists: boolean;
}

export type AdminKeysUpdateEvent = {
  channelId: string;
}

export enum ChannelStatus {
  SYNC_CREATED = 0,
  SYNC_UPDATED = 1,
  SYNC_DELETED = 2,
  SYNC_LOADED = 3
}

export type ChannelUpdateEvent = {
  channelId: string;
  status: ChannelStatus;
};

export type AllowList = Partial<Record<MessageType, Record<string, unknown>>>;

export type AllowLists = {
  allowWithTags: AllowList;
  allowWithoutTags: AllowList;
}

export enum ChannelNotificationLevel {
  NotifyNone = 10,
  NotifyPing = 20,
  NotifyAll = 40
}

export enum NotificationStatus {
  Mute = 0,
  WhenOpen = 1,
  Push = 2
}

export type NotificationState = {
  channelId: string;
  level: ChannelNotificationLevel;
  status: NotificationStatus;
}

export type NotificationUpdateEvent = {
  changedNotificationStates?: NotificationState[];
  deletedNotificationStates?: ChannelId[] | null;
}

export type DMNotificationLevelState = {
  pubkey: string;
  level: DMNotificationLevel;
}

export type DMNotificationsUpdateEvent = {
  changedNotificationStates: DMNotificationLevelState[];
  deletedNotificationStates: string[];
}

export type DMBlockedUserEvent = {
  pubkey: string;
  blocked: boolean;
}

export type ChannelDMTokenUpdate = {
  channelId: string;
  sendToken: boolean,
}

export enum DMNotificationLevel {
  NotifyNone = 10,
  NotifyAll = 40
}
