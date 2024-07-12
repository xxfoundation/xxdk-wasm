export enum ResponseTypes {
  SentMessage = "SentMessage",
  RoundResults = "RoundResults",
  QueryResponse = "QueryResponse"
}


export type RPCSend(
  cmixId: number,
  recipient: Uint8Array,
  pubkey: Uint8Array,
  request: Uint8Array,
  updateCallback: (json: Uint8Array) => void
) => Promise<Uint8Array);
