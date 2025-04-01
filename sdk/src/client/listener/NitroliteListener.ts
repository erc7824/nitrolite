import { 
  PublicClient,
  Address,
  Log,
  parseEventLogs
} from 'viem';

import { 
  ChannelOpenedEvent,
  ChannelClosedEvent,
  ChannelChallengedEvent,
  ChannelCheckpointedEvent,
  CustodyAbi
} from '../../abis';

import { Logger, defaultLogger } from '../../config';
import { Channel, ChannelId } from '../types';
import Errors from '../../errors';

export type EventCallback<T> = (data: T) => void | Promise<void>;

export interface ChannelOpenedData {
  channelId: ChannelId;
  channel: Channel;
}

export interface ChannelChallengedData {
  channelId: ChannelId;
  expiration: bigint;
}

export interface ChannelEventData {
  channelId: ChannelId;
}

/**
 * Listener for Nitrolite contract events
 */
export class NitroliteListener {
  private readonly publicClient: PublicClient;
  private readonly custodyAddress: Address;
  private readonly logger: Logger;
  private unwatch?: () => void;

  constructor(
    publicClient: PublicClient,
    custodyAddress: Address,
    logger: Logger = defaultLogger
  ) {
    if (!publicClient) {
      throw new Errors.MissingParameterError('publicClient');
    }
    if (!custodyAddress) {
      throw new Errors.MissingParameterError('custodyAddress');
    }

    this.publicClient = publicClient;
    this.custodyAddress = custodyAddress;
    this.logger = logger;
  }

  /**
   * Start listening to all channel events
   */
  public async listen(
    onChannelOpened?: EventCallback<ChannelOpenedData>,
    onChannelChallenged?: EventCallback<ChannelChallengedData>,
    onChannelCheckpointed?: EventCallback<ChannelEventData>,
    onChannelClosed?: EventCallback<ChannelEventData>
  ): Promise<void> {
    if (this.unwatch) {
      this.logger.warn('Listener is already running');
      return;
    }

    this.unwatch = await this.publicClient.watchContractEvent({
      address: this.custodyAddress,
      abi: CustodyAbi,
      onLogs: async (unparsedLogs: Log[]) => {
        const logs = parseEventLogs({
          abi: CustodyAbi,
          logs: unparsedLogs,
        })

        for (const log of logs) {
          try {
            switch (log.eventName) {
              case ChannelOpenedEvent:
                if (onChannelOpened) {
                  // TODO: call hooks after finalizing events format
                  console.log('Channel opened event:', log);
                  // await onChannelOpened({
                  //   channelId: log.args.channelId as ChannelId,
                  //   channel: log.args.channel as Channel
                  // });
                }
                break;
              
              case ChannelChallengedEvent:
                if (onChannelChallenged) {
                  // TODO: call hooks after finalizing events format
                  console.log('Channel challenged event:', log);
                  // await onChannelChallenged({
                  //   channelId: log.args.channelId as ChannelId,
                  //   expiration: log.args.expiration as bigint
                  // });
                }
                break;

              case ChannelCheckpointedEvent:
                if (onChannelCheckpointed) {
                  // TODO: call hooks after finalizing events format
                  console.log('Channel checkpointed event:', log);
                  // await onChannelCheckpointed({
                  //   channelId: log.args.channelId as ChannelId
                  // });
                }
                break;

              case ChannelClosedEvent:
                if (onChannelClosed) {
                  // TODO: call hooks after finalizing events format
                  console.log('Channel closed event:', log);

                  // await onChannelClosed({
                  //   channelId: log.args.channelId as ChannelId
                  // });
                }
                break;
            }
          } catch (error) {
            this.logger.error('Error processing event', {
              error,
              event: log.eventName
            });
          }
        }
      }
    });
  }

  /**
   * Stop listening to events
   */
  public stop(): void {
    if (this.unwatch) {
      this.unwatch();
      this.unwatch = undefined;
    }
  }
}
