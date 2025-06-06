import type { ApiResource } from '../types';
import type { SearchResultChannel } from 'types/api/search';

export const CHANNELS_API_RESOURCES = {
  search: {
    path: '/api/channels',
  },
} satisfies Record<string, ApiResource>;

export type ChannelsApiResourceName = `channels:${keyof typeof CHANNELS_API_RESOURCES}`;

/* eslint-disable @stylistic/indent */
export type ChannelsApiResourcePayload<R extends ChannelsApiResourceName> =
R extends 'channels:search' ? Array<SearchResultChannel> :
never;
/* eslint-enable @stylistic/indent */ 