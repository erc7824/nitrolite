import { Address } from 'viem';

/**
 * Contract addresses for Nitrolite infrastructure
 */
export interface ContractAddresses {
    /** Address of the Custody contract */
    custody: Address;

    /** Supported adjudicator addresses by type */
    adjudicator: Address;
}
