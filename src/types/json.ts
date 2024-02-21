export type ShareURLJSON = {
  url: string;
  password?: string;
}

export type IdentityJSON = {
  pubkey: string;
  codename: string;
  color: string;
  extension: string;
  codeset: number;
}

export type ChannelJSON = {
  receptionId?: string;
  channelId?: string;
  name: string;
  description: string;
}

export type VersionJSON = {
  current: string;
  updated: boolean;
  old: string;
}

export type IsReadyInfoJSON = {
  isReady: boolean;
  howClose: number;
}

export type MessageReceivedJSON = {
  uuid: number;
  channelId: string;
  update: boolean;
}