import { Address } from 'viem';

/**
 * Smart contract addresses for the Nitrolite framework
 * Replace these addresses with your deployed contract addresses
 */
export const CONTRACTS = {
  custody: '0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE' as Address,
  adjudicators: {
    // Base adjudicator for basic state validation
    base: '0x5fbdb2315678afecb367f032d93f642f64180aa3' as Address,
    
    // Custom adjudicator for counter application
    counter: '0x5fbdb2315678afecb367f032d93f642f64180aa3' as Address,
    
    // Additional adjudicators could be added here
    // numeric: '0x...' as Address,
    // sequential: '0x...' as Address,
  },
};

export default CONTRACTS;