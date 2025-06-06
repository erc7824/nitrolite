import { Box, Flex, Text } from '@chakra-ui/react';
import { useEffect, useState } from 'react';

import { Link } from 'toolkit/chakra/link';
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

function isLedgerEntry(item: unknown): item is LedgerEntry {
  if (!item || typeof item !== 'object') return false;
  const entry = item as Record<string, unknown>;
  return (
    typeof entry.id === 'number' &&
    typeof entry.accountId === 'string' &&
    typeof entry.accountType === 'number' &&
    typeof entry.asset === 'string' &&
    typeof entry.createdAt === 'string' &&
    typeof entry.credit === 'string' &&
    typeof entry.debit === 'string' &&
    typeof entry.participant === 'string'
  );
}

function isLedgerEntryArray(data: unknown): data is Array<LedgerEntry> {
  return Array.isArray(data) && data.every(isLedgerEntry);
}

const formatAmount = (amount: string): string => {
  return parseFloat(amount).toLocaleString(undefined, {
    minimumFractionDigits: 2,
    maximumFractionDigits: 8,
  });
};

export default function LatestLedgerEntries() {
  const [ entries, setEntries ] = useState<Array<LedgerEntry>>([]);
  const [ loading, setLoading ] = useState(true);
  const [ error, setError ] = useState<string | null>(null);

  useEffect(() => {
    const fetchLatestEntries = async() => {
      try {
        const response = await fetch('/api/latest-ledger-entries');
        if (!response.ok) {
          throw new Error('Failed to fetch latest ledger entries');
        }

        const data = await response.json();
        if (!isLedgerEntryArray(data)) {
          throw new Error('Invalid API response format');
        }

        setEntries(data);
        setError(null);
      } catch (err) {
        setError('Failed to load latest ledger entries');
        // eslint-disable-next-line no-console
        console.error('Error fetching latest ledger entries:', err);
      } finally {
        setLoading(false);
      }
    };

    fetchLatestEntries();
  }, []);

  if (loading) {
    return (
      <Box>
        {[ ...Array(5) ].map((_, i) => (
          <Box
            key={i}
            height="96px"
            borderRadius="md"
            mb={ i < 4 ? 4 : 0 }
          >
            <Box height="100%" bg="gray.100" _dark={{ bg: 'gray.700' }} borderRadius="md"/>
          </Box>
        ))}
      </Box>
    );
  }

  if (error) {
    return (
      <Text
        textAlign="center"
        py={ 4 }
        color="red.500"
      >
        {error}
      </Text>
    );
  }

  return (
    <Box>
      <Table>
        <Thead>
          <Tr>
            <Th>Account</Th>
            <Th>Asset</Th>
            <Th isNumeric>Credit</Th>
            <Th isNumeric>Debit</Th>
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
              <Td isNumeric color={parseFloat(entry.credit) > 0 ? 'green.500' : undefined}>
                {formatAmount(entry.credit)}
              </Td>
              <Td isNumeric color={parseFloat(entry.debit) > 0 ? 'red.500' : undefined}>
                {formatAmount(entry.debit)}
              </Td>
            </Tr>
          ))}
        </Tbody>
      </Table>
      <Flex justify="flex-end" mt={4}>
        <Link href="/ledger" color="blue.500">
          View all ledger entries →
        </Link>
      </Flex>
    </Box>
  );
} 