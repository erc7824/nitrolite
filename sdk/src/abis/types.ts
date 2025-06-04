import { Address } from 'viem';

/**
 * Contract addresses for Nitrolite infrastructure
 */
export interface ContractAddresses {
    /** Address of the Custody contract */
    custody: Address;

    /** Counterparty address for channel */
    guestAddress: Address;

    /** Address of token for channel */
    tokenAddress: Address;

    /** Supported adjudicator addresses by type */
    adjudicator: Address;
}
