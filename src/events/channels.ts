import { AdminKeysUpdateEvent, ChannelDMTokenUpdate, ChannelUpdateEvent,
         MessageDeletedEvent, MessageReceivedEvent, NicknameUpdatedEvent,
         NotificationUpdateEvent,
         UserMutedEvent } from '../types/events';
import { Decoder, adminKeysUpdateDecoder, channelDMTokenUpdateDecoder,
         channelUpdateEventDecoder, messageDeletedEventDecoder,
         messageReceivedEventDecoder, nicknameUpdatedEventDecoder,
         notificationUpdateEventDecoder,
         userMutedEventDecoder } from '../utils/decoders';

export enum ChannelEvents {
  NICKNAME_UPDATE = 1000,
  NOTIFICATION_UPDATE = 2000,
  MESSAGE_RECEIVED = 3000,
  USER_MUTED = 4000,
  MESSAGE_DELETED = 5000,
  ADMIN_KEY_UPDATE = 6000,
  DM_TOKEN_UPDATE = 7000,
  CHANNEL_UPDATE = 8000
}

export type ChannelEventMap = {
  [ChannelEvents.NICKNAME_UPDATE]: NicknameUpdatedEvent;
  [ChannelEvents.NOTIFICATION_UPDATE]: NotificationUpdateEvent;
  [ChannelEvents.MESSAGE_RECEIVED]: MessageReceivedEvent;
  [ChannelEvents.MESSAGE_DELETED]: MessageDeletedEvent;
  [ChannelEvents.USER_MUTED]: UserMutedEvent;
  [ChannelEvents.ADMIN_KEY_UPDATE]: AdminKeysUpdateEvent;
  [ChannelEvents.CHANNEL_UPDATE]: ChannelUpdateEvent[];
  [ChannelEvents.DM_TOKEN_UPDATE]: ChannelDMTokenUpdate;
}

export type ChannelEventHandlers = {
  [P in keyof ChannelEventMap]: (event: ChannelEventMap[P]) => void;
}

export type ChannelEventHandler = (eventType: ChannelEvents, data: unknown) => void;

// These cause a type error I haven't figured out:
// Type of computed property's value is '(thing: unknown) => unknown', which is not assignable to type 'Decoder...
// const channelsEventDecoderMap: { [P in keyof ChannelEventMap]: Decoder<ChannelEventMap[P]> } = {
//   [ChannelEvents.MESSAGE_RECEIVED]: messageReceivedEventDecoder,
//   [ChannelEvents.NOTIFICATION_UPDATE]: notificationUpdateEventDecoder,
//   [ChannelEvents.MESSAGE_DELETED]: messageDeletedEventDecoder,
//   [ChannelEvents.USER_MUTED]: userMutedEventDecoder,
//   [ChannelEvents.NICKNAME_UPDATE]: nicknameUpdatedEventDecoder,
//   [ChannelEvents.CHANNEL_UPDATE]: channelUpdateEventDecoder,
//   [ChannelEvents.DM_TOKEN_UPDATE]: channelDMTokenUpdateDecoder,
//   [ChannelEvents.ADMIN_KEY_UPDATE]: adminKeysUpdateDecoder,
// }
