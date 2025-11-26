import { Box, Flex, Text } from '@chakra-ui/react';
import { useEffect, useState } from 'react';

interface Channel {
  channelId: string;
  adjudicator: string;
  amount: string;
  chainId: number;
  challenge: number;
  createdAt: string;
  nonce: string;
  participant: string;
  status: string;
  token: string;
  updatedAt: string;
  version: number;
  wallet: string;
}

function isChannel(item: unknown): item is Channel {
  if (!item || typeof item !== 'object') return false;
  const channel = item as Record<string, unknown>;
  return (
    typeof channel.channelId === 'string' &&
    typeof channel.adjudicator === 'string' &&
    typeof channel.amount === 'string' &&
    typeof channel.chainId === 'number' &&
    typeof channel.challenge === 'number' &&
    typeof channel.createdAt === 'string' &&
    typeof channel.nonce === 'string' &&
    typeof channel.participant === 'string' &&
    typeof channel.status === 'string' &&
    typeof channel.token === 'string' &&
    typeof channel.updatedAt === 'string' &&
    typeof channel.version === 'number' &&
    typeof channel.wallet === 'string'
  );
}

function isChannelArray(data: unknown): data is Array<Channel> {
  return Array.isArray(data) && data.every(isChannel);
}

const getStatusColor = (status: string): string => {
  switch (status.toLowerCase()) {
    case 'open':
      return 'green.500';
    case 'closed':
      return 'red.500';
    case 'disputed':
      return 'yellow.500';
    default:
      return 'gray.500';
  }
};

export default function LatestChannels() {
  const [ channels, setChannels ] = useState<Array<Channel>>([]);
  const [ loading, setLoading ] = useState(true);
  const [ error, setError ] = useState<string | null>(null);

  useEffect(() => {
    const fetchLatestChannels = async() => {
      try {
        const response = await fetch('/api/latest-channels');
        if (!response.ok) {
          throw new Error('Failed to fetch latest channels');
        }

        const data = await response.json();
        if (!isChannelArray(data)) {
          throw new Error('Invalid API response format');
        }

        setChannels(data);
        setError(null);
      } catch (err) {
        setError('Failed to load latest channels');
        // eslint-disable-next-line no-console
        console.error('Error fetching latest channels:', err);
      } finally {
        setLoading(false);
      }
    };

    fetchLatestChannels();
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
      {channels.map((channel, index) => (
        <Box
          key={channel.channelId}
          bg="white"
          _dark={{ bg: 'gray.800' }}
          p={4}
          borderRadius="md"
          shadow="sm"
          _hover={{ shadow: 'md' }}
          transition="box-shadow 0.2s"
          width="100%"
          mb={index < channels.length - 1 ? 4 : 0}
        >
          <Box>
            <Flex justifyContent="space-between" alignItems="center" mb={2}>
              <Text
                fontSize="sm"
                fontWeight="medium"
                overflow="hidden"
                textOverflow="ellipsis"
                whiteSpace="nowrap"
              >
                Channel ID: {channel.channelId}
              </Text>
              <Text fontSize="sm" color="gray.500">
                Chain ID: {channel.chainId}
              </Text>
            </Flex>

            <Flex justifyContent="space-between" alignItems="center" mb={2}>
              <Text fontSize="xs" color="gray.500">
                Created: {new Date(channel.createdAt).toLocaleString()}
              </Text>
              <Text fontSize="xs" color="gray.500">
                Updated: {new Date(channel.updatedAt).toLocaleString()}
              </Text>
            </Flex>

            <Flex justifyContent="space-between" alignItems="center" mb={2}>
              <Text fontSize="sm">
                Status: <Text as="span" color={getStatusColor(channel.status)} fontWeight="medium">
                  {channel.status}
                </Text>
              </Text>
              <Text fontSize="sm" fontWeight="medium">
                Amount: {Number(channel.amount).toLocaleString()}
              </Text>
            </Flex>

            <Text
              fontSize="xs"
              color="gray.600"
              _dark={{ color: 'gray.300' }}
              mb={1}
              overflow="hidden"
              textOverflow="ellipsis"
              whiteSpace="nowrap"
            >
              Participant: {channel.participant}
            </Text>

            <Text
              fontSize="xs"
              color="gray.600"
              _dark={{ color: 'gray.300' }}
              overflow="hidden"
              textOverflow="ellipsis"
              whiteSpace="nowrap"
            >
              Wallet: {channel.wallet}
            </Text>
          </Box>
        </Box>
      ))}
    </Box>
  );
}
