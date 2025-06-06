import { Flex, Text } from '@chakra-ui/react';
import React from 'react';

import { route } from 'nextjs-routes';

import highlightText from 'lib/highlightText';
import { Link } from 'toolkit/chakra/link';

interface Props {
  data: {
    channelId: string;
    participant: string;
    wallet: string;
    amount: string;
    status: string;
  };
  searchTerm: string;
  onClick: (event: React.MouseEvent<HTMLAnchorElement>) => void;
}

const SearchBarSuggestChannel = ({ data, searchTerm, onClick }: Props) => {
  return (
    <Link
      href={route({ pathname: '/channels', query: { id: data.channelId } })}
      display="flex"
      flexDir="column"
      gap={2}
      py={3}
      px={1}
      borderBottomWidth="1px"
      borderColor="border.divider"
      _hover={{
        bgColor: { _light: 'blue.50', _dark: 'gray.800' },
      }}
      onClick={onClick}
      _last={{
        borderBottomWidth: '0',
      }}
    >
      <Flex alignItems="center" gap={2}>
        <Text fontWeight={700}>
          Channel ID: <span dangerouslySetInnerHTML={{ __html: highlightText(data.channelId, searchTerm) }}/>
        </Text>
        <Text
          px={2}
          py={0.5}
          borderRadius="full"
          fontSize="xs"
          bgColor={data.status === 'open' ? 'green.100' : 'gray.100'}
          color={data.status === 'open' ? 'green.800' : 'gray.800'}
        >
          {data.status}
        </Text>
      </Flex>
      <Flex direction="column" gap={1} color="text.secondary">
        <Text>
          Participant: <span dangerouslySetInnerHTML={{ __html: highlightText(data.participant, searchTerm) }}/>
        </Text>
        <Text>
          Wallet: <span dangerouslySetInnerHTML={{ __html: highlightText(data.wallet, searchTerm) }}/>
        </Text>
        <Text>Amount: {data.amount}</Text>
      </Flex>
    </Link>
  );
};

export default SearchBarSuggestChannel; 