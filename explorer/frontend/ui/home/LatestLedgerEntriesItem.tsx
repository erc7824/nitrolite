import { Box, Flex, Grid, Text } from '@chakra-ui/react';
import React from 'react';

import { Skeleton } from 'toolkit/chakra/skeleton';
import TimeAgoWithTooltip from 'ui/shared/TimeAgoWithTooltip';

interface LatestLedgerEntriesItemProps {
  entry: {
    id: number;
    accountId: string;
    accountType: number;
    asset: string;
    createdAt: string;
    credit: string;
    debit: string;
    participant: string;
  };
  isLoading?: boolean;
}

const formatAmount = (amount: string): string => {
  return parseFloat(amount).toLocaleString(undefined, {
    minimumFractionDigits: 2,
    maximumFractionDigits: 8,
  });
};

const LatestLedgerEntriesItem = ({ entry, isLoading }: LatestLedgerEntriesItemProps) => {
  return (
    <Box
      borderRadius="md"
      border="1px solid"
      borderColor="border.divider"
      p={3}
      _hover={{ bg: 'gray.50', _dark: { bg: 'gray.800' } }}
      transition="background-color 0.2s"
    >
      <Flex alignItems="center" overflow="hidden" w="100%" mb={3}>
        <Skeleton loading={isLoading} textStyle="xl" fontWeight={500} mr="auto">
          <Text>ID: {entry.id}</Text>
        </Skeleton>
        <TimeAgoWithTooltip
          timestamp={entry.createdAt}
          enableIncrement={!isLoading}
          isLoading={isLoading}
          color="text.secondary"
          display="inline-block"
          textStyle="sm"
          flexShrink={0}
          ml={2}
        />
      </Flex>
      <Grid gridGap={2} templateColumns="auto minmax(0, 1fr)" textStyle="sm">
        <Skeleton loading={isLoading}>Account ID</Skeleton>
        <Skeleton loading={isLoading} color="text.secondary" whiteSpace="nowrap" overflow="hidden" textOverflow="ellipsis">
          <Text fontFamily="mono">{entry.accountId}</Text>
        </Skeleton>

        <Skeleton loading={isLoading}>Asset</Skeleton>
        <Skeleton loading={isLoading} color="text.secondary">
          <Text>{entry.asset}</Text>
        </Skeleton>

        {Number(entry.credit) > 0 ? (
          <>
            <Skeleton loading={isLoading}>Credit</Skeleton>
            <Skeleton loading={isLoading} color="green.500">
              <Text>{formatAmount(entry.credit)}</Text>
            </Skeleton>
          </>
        ) : (
          <>
            <Skeleton loading={isLoading}>Debit</Skeleton>
            <Skeleton loading={isLoading} color="red.500">
              <Text>{formatAmount(entry.debit)}</Text>
            </Skeleton>
          </>
        )}
      </Grid>
    </Box>
  );
};

export default LatestLedgerEntriesItem;
