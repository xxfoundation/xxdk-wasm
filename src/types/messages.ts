import { MessageStatus, MessageType } from './db';
import { ChannelId } from './events';

// { [messageId]: { [emoji]: codename[] } }
export type ReactionInfo = { pubkey: string, codeset: number, id: string, status?: MessageStatus };
export type EmojiReactions =  Record<string, Record<string, ReactionInfo[]>>;

export interface Message {
  id: string;
  body: string;
  timestamp: string;
  repliedTo: string | null;
  type: MessageType;
  color?: string;
  codename: string;
  nickname?: string;
  plaintext: string | null;
  channelId: string;
  status?: MessageStatus;
  uuid: number;
  round: number;
  pubkey: string;
  pinned: boolean;
  hidden: boolean;
  codeset: number;
  dmToken?: number;
}

export type MessageUuid = Message['uuid'];
export type MessageId = Message['id'];

export type Contributor = Pick<Message, 'pubkey' | 'codeset' | 'codename' | 'nickname' | 'timestamp' | 'color' | 'dmToken'>;

export type MessagesState = {
  reactions: EmojiReactions;
  contributorsByChannelId: Record<ChannelId, Contributor[]>;
  byChannelId: Record<Message['channelId'], Record<MessageUuid, Message>>;
  sortedMessagesByChannelId: Record<Message['channelId'], Array<Message>>;
  commonChannelsByPubkey: Record<Message['pubkey'], ChannelId[]>;
  dmTokens: Record<Message['pubkey'], Message['dmToken']>;
};

