export const STATE_PATH = 'speakeasyapp';
export const CHANNELS_STORAGE_TAG = 'ELIXXIR_USERS_TAGS';
export const DMS_DATABASE_NAME = 'DMS_DATABASE_NAME';
export const ACCOUNT_SYNC = 'ACCOUNT_SYNC';
export const ACCOUNT_SYNC_SERVICE = 'ACCOUNT_SYNC_SERVICE';
export const CMIX_INITIALIZATION_KEY = 'cmixPreviouslyInitialized';

export const PIN_MESSAGE_LENGTH_MILLISECONDS = 1.814e+9
export const MAXIMUM_PAYLOAD_BLOCK_SIZE = 725;
export const MESSAGE_LEASE = 30000;
export const CMIX_NETWORK_READINESS_THRESHOLD = 0.1;

export const DUMMY_TRAFFIC_MAXIMUM_MESSAGES_PER_CYCLE = 3;
export const DUMMY_TRAFFIC_DURATION_TO_WAIT_BETWEEN_SENDS_MILLISECONDS = 15000;
export const DUMMY_TRAFFIC_UPPERBOUND_INTERVAL_BETWEEN_CYCLES_MILLISECONDS = 7000;
export const DUMMY_TRAFFIC_ARGS = [
  DUMMY_TRAFFIC_MAXIMUM_MESSAGES_PER_CYCLE,
  DUMMY_TRAFFIC_DURATION_TO_WAIT_BETWEEN_SENDS_MILLISECONDS,
  DUMMY_TRAFFIC_UPPERBOUND_INTERVAL_BETWEEN_CYCLES_MILLISECONDS
] as [number, number, number]

export const SLOW_MODE_TRACKING_PERIOD_MS = 5000;
export const FAST_MODE_TRACKING_PERIOD_MS = 1000;

export const AMOUNT_OF_IDENTITIES_TO_GENERATE = 20;
export const FOLLOWER_TIMEOUT_PERIOD = 50000;
export const MESSAGE_TAGS_LIMIT = 5;