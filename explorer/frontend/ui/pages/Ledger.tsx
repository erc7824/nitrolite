import { Box, Container, Flex, Text } from '@chakra-ui/react';
import React, { useEffect, useState, useCallback } from 'react';
import { FiSearch } from 'react-icons/fi';
import { useInView } from 'react-intersection-observer';

import { Heading } from 'toolkit/chakra/heading';
import { Input } from 'toolkit/chakra/input';
import { InputGroup } from 'toolkit/chakra/input-group';
import { Skeleton } from 'toolkit/chakra/skeleton';
import {
  TableRoot as Table,
  TableHeader as Thead,
  TableBody as Tbody,
  TableRow as Tr,
  TableColumnHeader as Th,
  TableCell as Td,
} from 'toolkit/chakra/table';

interface LedgerEntry {
  id: number;
  accountId: string;
  accountType: number;
  asset: string;
  createdAt: string;
  credit: string;
  debit: string;
  participant: string;
}

interface PaginatedResponse {
  entries: Array<LedgerEntry>;
  total: number;
  hasMore: boolean;
}

const formatAmount = (amount: string): string => {
  return parseFloat(amount).toLocaleString(undefined, {
    minimumFractionDigits: 2,
    maximumFractionDigits: 8,
  });
};

const LedgerPage = () => {
  const [entries, setEntries] = useState<Array<LedgerEntry>>([]);
  const [currentPage, setCurrentPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [debouncedSearchTerm, setDebouncedSearchTerm] = useState('');
  const { ref, inView } = useInView();

  // Debounce search term
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearchTerm(searchTerm);
    }, 300);

    return () => clearTimeout(timer);
  }, [searchTerm]);

  // Reset pagination when search term changes
  useEffect(() => {
    setCurrentPage(1);
    setEntries([]);
    fetchEntries(1, debouncedSearchTerm);
  }, [debouncedSearchTerm]);

  const fetchEntries = async(page: number, search: string = '') => {
    try {
      setLoading(true);
      setError(null);

      // Validate search term length
      if (search && search.length < 3) {
        setEntries([]);
        setHasMore(false);
        setError('Please enter at least 3 characters to search');
        return;
      }

      const response = await fetch(`/api/ledger-entries?page=${ page }&q=${ encodeURIComponent(search) }`);

      if (!response.ok) {
        throw new Error('Failed to fetch ledger entries');
      }

      const data = await response.json() as PaginatedResponse;

      if (page === 1) {
        setEntries(data.entries);
      } else {
        setEntries(prev => [...prev, ...data.entries]);
      }
      setHasMore(data.hasMore);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred while fetching entries');
      // eslint-disable-next-line no-console
      console.error('Error fetching ledger entries:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (inView && !loading && hasMore && !debouncedSearchTerm) {
      fetchEntries(currentPage + 1);
      setCurrentPage(prev => prev + 1);
    }
  }, [inView, loading, hasMore, currentPage, debouncedSearchTerm]);

  const handleSearchChange = useCallback((event: React.ChangeEvent<HTMLInputElement>) => {
    setSearchTerm(event.target.value);
  }, []);

  return (
    <Box as="main" py={8}>
      <Container maxW="container.xl">
        <Flex direction="column" gap={6}>
          <Flex justify="space-between" align="center">
            <Heading as="h1" size="xl">Ledger Entries</Heading>
          </Flex>

          <Box maxW="md">
            <InputGroup
              startElement={(
                <Box color="text.secondary">
                  <FiSearch/>
                </Box>
              )}
            >
              <Input
                placeholder="Search by account ID, participant, or asset"
                value={searchTerm}
                onChange={handleSearchChange}
                variant="outline"
              />
            </InputGroup>
          </Box>

          {error ? (
            <Text color="red.500" textAlign="center">{error}</Text>
          ) : (
            <Table>
              <Thead>
                <Tr>
                  <Th>Account ID</Th>
                  <Th>Asset</Th>
                  <Th>Participant</Th>
                  <Th isNumeric>Credit</Th>
                  <Th isNumeric>Debit</Th>
                  <Th isNumeric>Created At</Th>
                </Tr>
              </Thead>
              <Tbody>
                {entries.map((entry) => (
                  <Tr
                    key={entry.id}
                    _hover={{ bg: 'gray.50', _dark: { bg: 'gray.800' } }}
                    transition="background-color 0.2s"
                  >
                    <Td fontFamily="mono">{entry.accountId}</Td>
                    <Td>{entry.asset}</Td>
                    <Td fontFamily="mono">{entry.participant}</Td>
                    <Td isNumeric color={parseFloat(entry.credit) > 0 ? 'green.500' : undefined}>
                      {formatAmount(entry.credit)}
                    </Td>
                    <Td isNumeric color={parseFloat(entry.debit) > 0 ? 'red.500' : undefined}>
                      {formatAmount(entry.debit)}
                    </Td>
                    <Td isNumeric>
                      {new Date(entry.createdAt).toLocaleDateString()}
                    </Td>
                  </Tr>
                ))}
              </Tbody>
            </Table>
          )}

          {/* Loading and Load More section */}
          <Box ref={ref} pt={4}>
            {loading && (
              <Flex justify="center" py={4}>
                <Skeleton height="40px" width="200px" loading={true}/>
              </Flex>
            )}
            {!loading && !hasMore && entries.length > 0 && (
              <Text textAlign="center" color="text.secondary">
                You've reached the end of the list
              </Text>
            )}
          </Box>
        </Flex>
      </Container>
    </Box>
  );
};

export default LedgerPage; 