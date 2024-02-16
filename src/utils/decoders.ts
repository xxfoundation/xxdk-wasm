import {
  AdminKeysUpdateEvent,
  ChannelId,
  ChannelStatus,
  ChannelUpdateEvent,
  DMBlockedUserEvent,
  DMNotificationLevelState,
  DMNotificationsUpdateEvent,
  DMReceivedEvent,
  MessageDeletedEvent,
  MessageReceivedEvent,
  NicknameUpdatedEvent,
  ChannelNotificationLevel,
  NotificationState,
  NotificationStatus,
  NotificationUpdateEvent,
  UserMutedEvent,
  DMNotificationLevel,
  ChannelDMTokenUpdate
} from '../types/events';

import {
  ChannelJSON,
  IdentityJSON,
  IsReadyInfoJSON,
  ShareURLJSON,
  VersionJSON
} from '../types/json';

import { KVEntry } from '../types/collective';
import { Err, JsonDecoder } from 'ts.data.json';
import { decoder as uintDecoder } from './index';


const attemptParse = (object: unknown) => {
  let parsed = object;
  if (typeof object === 'string') {
    try {
      parsed = JSON.parse(object);
    } catch (e) {
      console.error('Failed to parse string in decoder', object);
    }
  }
  return parsed;
}

export const makeDecoder = <T>(decoder: JsonDecoder.Decoder<T>) => (thing: unknown): T => {
  const object = thing instanceof Uint8Array ? uintDecoder.decode(thing) : thing;
  const parsed = typeof object === 'string' ? attemptParse(object) : object;
  const result = decoder.decode(parsed);
  if (result instanceof Err) {
    throw new Error(`Unexpected JSON: ${JSON.stringify(parsed)}, Error: ${result.error}`);
  } else {
    return result.value;
  }
}

export type Decoder<T> = ReturnType<typeof makeDecoder<T>>;

const uint8ArrayDecoder = JsonDecoder.array(JsonDecoder.number, 'Uint8Decoder');

export const channelDecoder = makeDecoder(JsonDecoder.object<ChannelJSON>(
  {
    receptionId: JsonDecoder.optional(JsonDecoder.string),
    channelId: JsonDecoder.optional(JsonDecoder.string),
    name: JsonDecoder.string,
    description: JsonDecoder.string,
  },
  'ChannelDecoder',
  {
    receptionId: 'ReceptionID',
    channelId: 'ChannelID',
    name: 'Name',
    description: 'Description'
  }
));

export const identityDecoder = makeDecoder(JsonDecoder.object<IdentityJSON>(
  {
    pubkey: JsonDecoder.string,
    codename: JsonDecoder.string,
    color: JsonDecoder.string,
    extension: JsonDecoder.string,
    codeset: JsonDecoder.number
  },
  'IdentityDecoder',
  {
    pubkey: 'PubKey',
    codename: 'Codename',
    color: 'Color',
    extension: 'Extension',
    codeset: 'CodesetVersion'
  }
));

export const shareUrlDecoder = makeDecoder(JsonDecoder.object<ShareURLJSON>(
  {
    password: JsonDecoder.optional(JsonDecoder.string),
    url: JsonDecoder.string
  },
  'ShareUrlDecoder'
));

export const isReadyInfoDecoder = makeDecoder(JsonDecoder.object<IsReadyInfoJSON>({
  isReady: JsonDecoder.boolean,
  howClose: JsonDecoder.number
}, 'IsReadyInfoDecoder', {
  isReady: 'IsReady',
  howClose: 'HowClose'
}))

export const pubkeyArrayDecoder = makeDecoder(JsonDecoder.array<string>(JsonDecoder.string, 'PubkeyArrayDecoder'));

export const versionDecoder = makeDecoder(JsonDecoder.object<VersionJSON>(
  {
    current: JsonDecoder.string,
    updated: JsonDecoder.boolean,
    old: JsonDecoder.string
  },
  'VersionDecoder'
));

export const kvEntryDecoder = makeDecoder(JsonDecoder.nullable(JsonDecoder.object<KVEntry>(
  {
    data: JsonDecoder.string,
    version: JsonDecoder.number,
    timestamp: JsonDecoder.string,
  }, 'KVEntryDecoder', {
    data: 'Data',
    version: 'Version',
    timestamp: 'Timestamp'
  }
)));

export const messageReceivedEventDecoder = makeDecoder(JsonDecoder.object<MessageReceivedEvent>(
  {
    uuid: JsonDecoder.number,
    channelId: JsonDecoder.string,
    update: JsonDecoder.boolean,
  },
  'MessageReceivedEventDecoder',
  {
    channelId: 'channelID',
  }
));

export const userMutedEventDecoder = makeDecoder(JsonDecoder.object<UserMutedEvent>(
  {
    channelId: JsonDecoder.string,
    pubkey: JsonDecoder.string,
    unmute: JsonDecoder.boolean,
  },
  'UserMutedEventDecoder',
  {
    pubkey: 'pubKey',
    channelId: 'channelID',
    unmute: 'unmute'
  }
));

const messageIdDecoder = uint8ArrayDecoder.map((s) => Buffer.from(s).toString('base64'))

export const messageDeletedEventDecoder = makeDecoder(JsonDecoder.object<MessageDeletedEvent>(
  {
    messageId: messageIdDecoder,
  },
  'MessageDeletedDecoder',
  {
    messageId: 'MessageId'
  }
));

export const nicknameUpdatedEventDecoder = makeDecoder(JsonDecoder.object<NicknameUpdatedEvent>(
  {
    channelId: JsonDecoder.string,
    nickname: JsonDecoder.string,
    exists: JsonDecoder.boolean
  },
  'NicknameUpdatedEventDecoder',
  {
    channelId: 'channelID',
  }
));

export const notificationLevelDecoder = JsonDecoder.enumeration<ChannelNotificationLevel>(ChannelNotificationLevel, 'NotificationLevelDecoder');
export const notificationStatusDecoder = JsonDecoder.enumeration<NotificationStatus>(NotificationStatus, 'NotificationStatusDecoder');
const notificationStateDecoder = JsonDecoder.object<NotificationState>(
  {
    channelId: JsonDecoder.string,
    level: notificationLevelDecoder,
    status: notificationStatusDecoder
  },
  'NotificationStateDecoder',
  {
    channelId: 'channelID',
  }
);

export const notificationUpdateEventDecoder = makeDecoder(JsonDecoder.object<NotificationUpdateEvent>(
  {
    changedNotificationStates: JsonDecoder.array<NotificationState>(notificationStateDecoder, 'ChangedNotificationStatesDecoder'),
    deletedNotificationStates: JsonDecoder.nullable(JsonDecoder.array<ChannelId>(JsonDecoder.string, 'DeletedNotificationStatesDecoder')),
  },
  'NotificationUpdateEventDecoder',
))

export const channelFavoritesDecoder = makeDecoder(JsonDecoder.array<string>(JsonDecoder.string, 'ChannelFavoritesDecoder'))

export const adminKeysUpdateDecoder = makeDecoder(JsonDecoder.object<AdminKeysUpdateEvent>(
  {
    channelId: JsonDecoder.string,
  },
  'AdminKeysUpdateDecoder',
  {
    channelId: 'channelID',
  }
));

export const channelDMTokenUpdateDecoder = makeDecoder(JsonDecoder.object<ChannelDMTokenUpdate>(
  {
    channelId: JsonDecoder.string,
    sendToken: JsonDecoder.boolean
  },
  'ChannelDMTokenDecoder',
  {
    channelId: 'channelID'
  }
))

export const channelStatusDecoder = JsonDecoder.enumeration<ChannelStatus>(ChannelStatus, 'ChannelStatusDecoder');

export const channelUpdateEventDecoder = makeDecoder(
  JsonDecoder.array<ChannelUpdateEvent>(
    JsonDecoder.object<ChannelUpdateEvent>(
      {
        channelId: JsonDecoder.string,
        status: channelStatusDecoder,
      },
      'ChannelUpdateDecoder',
      {
        channelId: 'channelID',
      }
    ),
    'ChannelUpdateEventDecoder')
);

export const dmNotificationLevelDecoder = JsonDecoder.enumeration<DMNotificationLevel>(DMNotificationLevel, 'DMNotificationLevel');

const dmNotificationLevelStatesDecoder = JsonDecoder.array<DMNotificationLevelState>(
  JsonDecoder.object<DMNotificationLevelState>(
    {
      pubkey: JsonDecoder.string,
      level: dmNotificationLevelDecoder
    },
    'DMNotificationLevelState',
    {
      pubkey: 'pubKey',
    }
  ),
  'DMNotificationLevelStateArrayDecoder'
);

const dmDeletedNotificationStatesDecoder = JsonDecoder.array<string>(JsonDecoder.string, 'DmDeletedNotificationStateArray');

export const dmNotificationsUpdateEventDecoder = makeDecoder(JsonDecoder.object<DMNotificationsUpdateEvent>(
  {
    changedNotificationStates: dmNotificationLevelStatesDecoder,
    deletedNotificationStates: dmDeletedNotificationStatesDecoder,
  },
  'DMNotificationsUpdateEventDecoder',
  {
    changedNotificationStates: 'changed',
    deletedNotificationStates: 'deleted'
  }
));

export const blockedUserEventDecoder = makeDecoder(JsonDecoder.object<DMBlockedUserEvent>(
  {
    pubkey: JsonDecoder.string,
    blocked: JsonDecoder.boolean
  },
  'DMBlockedUserEvent',
  {
    pubkey: 'user',
  }
));

export const dmReceivedEventDecoder = makeDecoder(JsonDecoder.object<DMReceivedEvent>(
  {
    uuid: JsonDecoder.number,
    pubkey: JsonDecoder.string,
    update: JsonDecoder.boolean,
    conversationUpdated: JsonDecoder.boolean
  },
  'DMReceivedEventDecoder',
  {
    pubkey: 'pubKey',
    update: 'messageUpdate',
    conversationUpdated: 'conversationUpdate'
  }
))
