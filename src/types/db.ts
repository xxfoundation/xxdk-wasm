export enum MessageType {
  Text = 1,
  Reply = 2,
  Reaction = 3,
  Silent = 4,
  Delete = 101,
  Pinned = 102,
  Mute = 103,
  AdminReplay = 104,
  FileTransfer = 40000
}

export enum MessageStatus {
  Unsent    = 0,
  Sent      = 1,
  Delivered = 2,
  Failed    = 3
}

export type DBMessage = {
  id: number;
  nickname: string;
  message_id: string;
  channel_id: string;
  parent_message_id: null | string;
  timestamp: string;
  lease: number;
  status: MessageStatus;
  hidden: boolean,
  pinned: boolean;
  text: string;
  type: MessageType;
  round: number;
  pubkey: string;
  codeset_version: number;
  dm_token: number;
}

export type DBChannel = {
  id: string;
  name: string;
  description: string;
}

export type DBDirectMessage = {
  id: number;
  message_id: string;
  conversation_pub_key: string;
  parent_message_id: string;
  codeset_version: number;
  sender_pub_key: string;
  timestamp: string;
  status: MessageStatus;
  text: string;
  type: MessageType;
  round: number;
}

export type DBConversation = {
  pub_key: string;
  nickname?: string;
  token: number; 
  codeset_version: number;
  blocked: boolean;
}
