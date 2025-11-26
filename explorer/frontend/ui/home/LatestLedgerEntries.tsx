import { Box, Flex, Text, VStack } from '@chakra-ui/react';
import React, { useEffect, useState } from 'react';
import { Heading } from 'toolkit/chakra/heading';
import { Link } from 'toolkit/chakra/link';
import { Skeleton } from 'toolkit/chakra/skeleton'; // For loading state
import LatestLedgerEntriesItem from './LatestLedgerEntriesItem';

// Define a type matching the API response
interface LedgerEntryAPI {
  id: number;
  accountId: string;
  accountType: number;
  asset: string;
  createdAt: string;
  credit: string;
  debit: string;
  participant: string;
}

const LatestLedgerEntries = () => {
  const [entries, setEntries] = useState<Array<LedgerEntryAPI>>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isError, setIsError] = useState(false);
  const entriesMaxCount = 5; // Show 5 entries

  useEffect(() => {
    const fetchEntries = async () => {
      setIsLoading(true);
      setIsError(false);
      try {
        const response = await fetch('/api/latest-ledger-entries');
        if (!response.ok) {
          throw new Error('Failed to fetch latest entries');
        }
        const data = await response.json();
        if (Array.isArray(data)) {
          setEntries(data);
        } else {
          setEntries([]);
        }
      } catch (error) {
        setIsError(true);
      } finally {
        setIsLoading(false);
      }
    };
    fetchEntries();
  }, []);

  return (
    <Box>
      <Flex justify="space-between" align="center" mb={4}>
        <Heading size="md">Latest Ledger Entries</Heading>
        <Link href="/ledger" color="blue.500">
          View all →
        </Link>
      </Flex>
      {isError ? (
        <Text color="red.500">Failed to load latest entries.</Text>
      ) : (
        <VStack align="stretch">
          {isLoading
            ? Array.from({ length: entriesMaxCount }).map((_, idx) => (
                <Skeleton key={idx} height="40px" loading={true} />
              ))
            : entries.length > 0
            ? entries.map((entry) => (
                <LatestLedgerEntriesItem key={entry.id} entry={entry} />
              ))
            : <Text>No data. Please reload the page.</Text>
          }
        </VStack>
      )}
    </Box>
  );
};

export default LatestLedgerEntries; 