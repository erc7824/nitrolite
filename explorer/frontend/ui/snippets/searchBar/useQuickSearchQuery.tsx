import React from 'react';

import { isBech32Address, fromBech32Address } from 'lib/address/bech32';
import useApiQuery from 'lib/api/useApiQuery';
import useDebounce from 'lib/hooks/useDebounce';

export default function useQuickSearchQuery() {
  const [ searchTerm, setSearchTerm ] = React.useState('');

  const debouncedSearchTerm = useDebounce(searchTerm, 300);

  const channelQuery = useApiQuery('channels:search', {
    queryParams: { q: debouncedSearchTerm },
    queryOptions: { enabled: debouncedSearchTerm.trim().length > 0 },
  });

  const query = useApiQuery('general:quick_search', {
    queryParams: { q: isBech32Address(debouncedSearchTerm) ? fromBech32Address(debouncedSearchTerm) : debouncedSearchTerm },
    queryOptions: { enabled: debouncedSearchTerm.trim().length > 0 },
  });

  const redirectCheckQuery = useApiQuery('general:search_check_redirect', {
    queryParams: { q: debouncedSearchTerm },
    queryOptions: { enabled: Boolean(debouncedSearchTerm) },
  });

  const combinedQuery = React.useMemo(() => {
    if (!query.data && !channelQuery.data) {
      return query;
    }

    return {
      ...query,
      data: [
        ...(channelQuery.data || []),
        ...(query.data || []),
      ],
    };
  }, [query, channelQuery]);

  return React.useMemo(() => ({
    searchTerm,
    debouncedSearchTerm,
    handleSearchTermChange: setSearchTerm,
    query: combinedQuery,
    redirectCheckQuery,
  }), [ debouncedSearchTerm, combinedQuery, redirectCheckQuery, searchTerm ]);
}
