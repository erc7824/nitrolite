import { Box, Flex } from '@chakra-ui/react';
import React from 'react';

import { Heading } from 'toolkit/chakra/heading';
import { Link } from 'toolkit/chakra/link';

import LatestChannelsComponent from '../../components/LatestChannels';

const LatestChannels = () => {
  return (
    <Box
      backgroundColor="white"
      _dark={{ backgroundColor: 'gray.800' }}
      padding={ 6 }
      borderRadius="md"
      width={{ base: '100%', lg: '400px' }}
      height="fit-content"
      flexShrink={ 0 }
    >
      <Flex justifyContent="space-between" alignItems="center" mb={ 6 }>
        <Heading size="md" mb={ 0 }>Latest channels</Heading>
        <Link href="/channels" fontSize="sm">View all</Link>
      </Flex>
      <Box>
        <LatestChannelsComponent/>
      </Box>
    </Box>
  );
};

export default LatestChannels;
