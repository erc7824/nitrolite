import { Box, Container, Flex, Grid, Text, Spinner, Center, Input } from '@chakra-ui/react';
import type { Channel } from '@prisma/client';
import React, { useEffect, useState, useCallback } from 'react';
import { FiSearch } from 'react-icons/fi';
import { useInView } from 'react-intersection-observer';

import { Badge } from 'toolkit/chakra/badge';
import { Button } from 'toolkit/chakra/button';
import { Heading } from 'toolkit/chakra/heading';
import { InputGroup } from 'toolkit/chakra/input-group';

interface PaginatedResponse {
  channels: Array<Channel>;
  total: number;
  hasMore: boolean;
}

interface SearchResult {
  type: 'channel';
  data: {
    channelId: string;
    participant: string;
    wallet: string;
    amount: string;
    status: string;
    adjudicator: string;
    chainId: number;
    challenge: number;
    nonce: string;
    token: string;
    version: number;
    createdAt: string;
    updatedAt: string;
  };
}

export default function ChannelsPage() {
  const [channels, setChannels] = useState<Array<Channel>>([]);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [totalChannels, setTotalChannels] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isInitialLoad, setIsInitialLoad] = useState(true);
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
    setChannels([]);
    fetchChannels(1, debouncedSearchTerm);
  }, [debouncedSearchTerm]);

  const fetchChannels = async(page: number, search: string = '') => {
    try {
      setLoading(true);
      setError(null);

      // Validate search term length
      if (search && search.length < 3) {
        setChannels([]);
        setTotalPages(1);
        setCurrentPage(1);
        setTotalChannels(0);
        setError('Please enter at least 3 characters to search');
        return;
      }

      const response = await fetch(`/api/channels?page=${ page }&q=${ encodeURIComponent(search) }`);
      
      if (!response.ok) {
        const errorData = await response.json().catch(() => ({ error: 'Failed to fetch channels' }));
        throw new Error(errorData.error || 'Failed to fetch channels');
      }

      const data = await response.json() as PaginatedResponse | Array<SearchResult>;

      // Validate response data
      if (!data || (Array.isArray(data) && !data.length && search) || (!Array.isArray(data) && !data.channels)) {
        setChannels([]);
        setTotalPages(1);
        setCurrentPage(1);
        setTotalChannels(0);
        return;
      }

      // Handle search results format
      if (search) {
        const searchResults = data as Array<SearchResult>;
        try {
          setChannels(searchResults.map((item) => ({
            channelId: item.data.channelId,
            participant: item.data.participant,
            wallet: item.data.wallet,
            amount: BigInt(item.data.amount || '0'),
            status: item.data.status,
            adjudicator: item.data.adjudicator,
            chainId: item.data.chainId,
            challenge: item.data.challenge,
            nonce: BigInt(item.data.nonce || '0'),
            token: item.data.token,
            version: item.data.version,
            createdAt: new Date(item.data.createdAt || Date.now()),
            updatedAt: new Date(item.data.updatedAt || Date.now()),
          })));
          setTotalPages(1); // Search results come in a single page
          setCurrentPage(1);
          setTotalChannels(searchResults.length);
        } catch (err) {
          // eslint-disable-next-line no-console
          console.error('Error processing search results:', err);
          setError('Invalid data received from server');
          setChannels([]);
          setTotalPages(1);
          setCurrentPage(1);
          setTotalChannels(0);
        }
      } else {
        // Handle regular paginated response
        const paginatedData = data as PaginatedResponse;
        if (page === 1) {
          setChannels(paginatedData.channels);
        } else {
          setChannels(prev => [ ...prev, ...paginatedData.channels ]);
        }
        setTotalPages(Math.ceil(paginatedData.total / 10)); // 10 is ITEMS_PER_PAGE from API
        setCurrentPage(page);
        setTotalChannels(paginatedData.total);
      }
    } catch (error) {
      setError(error instanceof Error ? error.message : 'An error occurred while fetching channels');
      // eslint-disable-next-line no-console
      console.error('Error fetching channels:', error);
      setChannels([]);
      setTotalPages(1);
      setCurrentPage(1);
      setTotalChannels(0);
    } finally {
      setLoading(false);
      setIsInitialLoad(false);
    }
  };

  useEffect(() => {
    if (inView && !loading && currentPage < totalPages && !debouncedSearchTerm) {
      // Only load more pages when not searching
      fetchChannels(currentPage + 1, debouncedSearchTerm);
    }
  }, [ inView, loading, currentPage, totalPages, debouncedSearchTerm ]);

  const formatBigInt = (value: bigint): string => {
    return value.toString();
  };

  const getStatusColor = (status: string): string => {
    switch (status) {
      case 'open':
        return 'green.500';
      case 'disputed':
        return 'yellow.500';
      case 'closed':
        return 'red.500';
      default:
        return 'gray.500';
    }
  };

  const handleLoadMore = useCallback(() => {
    if (!loading && currentPage < totalPages) {
      fetchChannels(currentPage + 1, debouncedSearchTerm);
    }
  }, [loading, currentPage, totalPages, debouncedSearchTerm]);

  const handleRetry = useCallback(() => {
    fetchChannels(1, debouncedSearchTerm);
  }, [debouncedSearchTerm]);

  const handleSearchChange = useCallback((event: React.ChangeEvent<HTMLInputElement>) => {
    setSearchTerm(event.target.value);
  }, []);

  if (isInitialLoad) {
    return (
      <Center minH="60vh">
        <Flex direction="column" align="center" gap={4}>
          <Spinner size="xl" />
          <Text>Loading channels...</Text>
        </Flex>
      </Center>
    );
  }

  if (error) {
    return (
      <Center minH="60vh">
        <Flex direction="column" align="center" gap={4}>
          <Text color="red.500">{error}</Text>
          <Button onClick={handleRetry} colorScheme="blue">
            Try Again
          </Button>
        </Flex>
      </Center>
    );
  }

  return (
    <Box as="main">
      <Container maxW="container.xl" py={8}>
        <Flex direction="column" gap={6}>
          <Flex justify="space-between" align="center">
            <Heading as="h1" size="xl">Channels</Heading>
            <Text color="text.secondary">
              {loading ? (
                'Loading...'
              ) : (
                `${ totalChannels?.toLocaleString() ?? 0 } total channel${ totalChannels !== 1 ? 's' : '' }`
              )}
            </Text>
          </Flex>

          <Box maxW="md">
            <InputGroup
              startElement={(
                <Box color="text.secondary">
                  <FiSearch />
                </Box>
              )}
            >
              <Input
                placeholder="Search by channel ID, participant, or wallet address"
                value={searchTerm}
                onChange={handleSearchChange}
                variant="outline"
              />
            </InputGroup>
          </Box>

          {!loading && (channels?.length ?? 0) === 0 && (
            <Center py={8}>
              <Text color="text.secondary">
                {debouncedSearchTerm ? 'No channels found matching your search' : 'No channels available'}
              </Text>
            </Center>
          )}

          <Grid
            templateColumns={{ base: '1fr', md: 'repeat(2, 1fr)', lg: 'repeat(3, 1fr)' }}
            gap={6}
          >
            {channels?.map((channel) => (
              <Box
                key={channel.channelId}
                borderWidth="1px"
                borderRadius="lg"
                overflow="hidden"
                p={5}
                transition="all 0.2s"
                _hover={{ shadow: 'lg', transform: 'translateY(-2px)' }}
                bg={{ _light: 'white', _dark: 'gray.800' }}
              >
                <Flex justify="space-between" align="center" mb={4}>
                  <Text fontWeight="bold" fontSize="md" maxW="70%" overflow="hidden" textOverflow="ellipsis" whiteSpace="nowrap">
                    Channel ID: {channel.channelId.substring(0, 10)}...
                  </Text>
                  <Badge
                    variant="subtle"
                    color={getStatusColor(channel.status)}
                    bg={`${ getStatusColor(channel.status) }.100`}
                    px={2}
                    py={1}
                    borderRadius="md"
                  >
                    {channel.status}
                  </Badge>
                </Flex>

                <Box color="text.secondary">
                  <Flex direction="column" gap={2}>
                    <Flex justify="space-between">
                      <Text fontSize="sm" color="text.secondary">Participant</Text>
                      <Text fontSize="sm" fontFamily="mono">
                        {channel.participant.substring(0, 10)}...
                      </Text>
                    </Flex>
                    <Flex justify="space-between">
                      <Text fontSize="sm" color="text.secondary">Amount</Text>
                      <Text fontSize="sm" fontFamily="mono">
                        {formatBigInt(channel.amount)}
                      </Text>
                    </Flex>
                    <Flex justify="space-between">
                      <Text fontSize="sm" color="text.secondary">Token</Text>
                      <Text fontSize="sm" fontFamily="mono">
                        {channel.token.substring(0, 10)}...
                      </Text>
                    </Flex>
                    <Flex justify="space-between">
                      <Text fontSize="sm" color="text.secondary">Created</Text>
                      <Text fontSize="sm">
                        {new Date(channel.createdAt).toLocaleDateString()}
                      </Text>
                    </Flex>
                  </Flex>
                </Box>
              </Box>
            ))}
          </Grid>

          {/* Loading and Load More section */}
          <Box ref={ref} pt={4}>
            {loading && (
              <Flex justify="center" py={4}>
                <Spinner size="lg" />
              </Flex>
            )}
            {!loading && currentPage < totalPages && (
              <Flex justify="center">
                <Button
                  onClick={handleLoadMore}
                  size="lg"
                  variant="outline"
                  colorScheme="blue"
                  loading={loading}
                >
                  Load More Channels
                </Button>
              </Flex>
            )}
            {currentPage === totalPages && channels.length > 0 && (
              <Text textAlign="center" color="text.secondary" mt={4}>
                You've reached the end of the list
              </Text>
            )}
          </Box>
        </Flex>
      </Container>
    </Box>
  );
}
