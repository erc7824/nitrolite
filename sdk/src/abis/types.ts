import { Address } from 'viem';

/**
 * Contract addresses for Nitrolite infrastructure
 */
export interface ContractAddresses {
  /** Address of the Custody contract */
  custody: Address;
  
  /** Supported adjudicator addresses by type */
  adjudicators: {
    /** Base adjudicator address */
    base?: Address;
    
    /** Numeric adjudicator address */
    numeric?: Address;
    
    /** Sequential adjudicator address */
    sequential?: Address;
    
    /** Other custom adjudicators */
    [key: string]: Address | undefined;
  };
}

/**
 * Configuration for ABI usage
 */
export interface AbiConfig {
  /** Chain ID the ABIs are for */
  chainId: number;
  
  /** Contract addresses */
  addresses: ContractAddresses;
}