import { Box, Text, chakra } from '@chakra-ui/react';
import React, { useEffect } from 'react';

import { Image } from 'toolkit/chakra/image';
import { Link } from 'toolkit/chakra/link';
import { Skeleton } from 'toolkit/chakra/skeleton';
import { ndash } from 'toolkit/utils/htmlEntities';
import { isBrowser } from 'toolkit/utils/isBrowser';

type AdData = {
  ad: {
    name: string;
    description_short: string;
    thumbnail: string;
    url: string;
    cta_button: string;
    impressionUrl?: string;
  };
};

// const MOCK: AdData = {
//   ad: {
//     url: 'https://unsplash.com/s/photos/cute-kitten',
//     thumbnail: 'https://raw.githubusercontent.com/blockscout/frontend-configs/main/configs/network-icons/gnosis.svg',
//     name: 'All about kitties',
//     description_short: 'To see millions picture of cute kitties',
//     cta_button: 'click here',
//   },
// };

const FETCH_TIMEOUT = 5000; // 5 seconds timeout
const AD_DISABLED_KEY = 'coinzilla_ad_disabled';
const AD_DISABLED_DURATION = 5 * 60 * 1000; // 5 minutes

const shouldDisableAds = () => {
  try {
    const disabledUntil = localStorage.getItem(AD_DISABLED_KEY);
    if (disabledUntil && Number(disabledUntil) > Date.now()) {
      return true;
    }
    localStorage.removeItem(AD_DISABLED_KEY);
    return false;
  } catch {
    return false;
  }
};

const disableAdsTemporarily = () => {
  try {
    localStorage.setItem(AD_DISABLED_KEY, String(Date.now() + AD_DISABLED_DURATION));
  } catch {
    // Ignore storage errors
  }
};

const CoinzillaTextAd = ({ className }: { className?: string }) => {
  const [ adData, setAdData ] = React.useState<AdData | null>(null);
  const [ isLoading, setIsLoading ] = React.useState(true);

  useEffect(() => {
    let isMounted = true;

    const fetchWithTimeout = async(url: string, options: RequestInit = {}) => {
      const controller = new AbortController();
      const id = setTimeout(() => controller.abort(), FETCH_TIMEOUT);

      try {
        const response = await fetch(url, {
          ...options,
          signal: controller.signal,
        });
        clearTimeout(id);
        return response;
      } catch (error) {
        clearTimeout(id);
        throw error;
      }
    };

    const loadAd = async() => {
      if (!isBrowser() || shouldDisableAds()) {
        setIsLoading(false);
        return;
      }

      try {
        // Check network connectivity first
        if (!navigator.onLine) {
          throw new Error('No network connection');
        }

        const response = await fetchWithTimeout('https://request-global.czilladx.com/serve/native.php?z=19260bf627546ab7242');
        if (!isMounted) return;

        if (!response.ok) {
          throw new Error(`HTTP error! status: ${ response.status }`);
        }

        const data = await response.json() as AdData;
        if (!isMounted) return;

        if (data?.ad) {
          setAdData(data);
          if (data.ad.impressionUrl) {
            // Fire and forget impression tracking
            fetchWithTimeout(data.ad.impressionUrl).catch(() => { /* Silently handle impression tracking errors */ });
          }
        }
      } catch (error) {
        // eslint-disable-next-line no-console
        console.error('Failed to fetch ad:', error);
        if (isMounted) {
          setAdData(null);
          // Disable ads temporarily if we encounter network errors
          disableAdsTemporarily();
        }
      } finally {
        if (isMounted) {
          setIsLoading(false);
        }
      }
    };

    loadAd();

    return () => {
      isMounted = false;
    };
  }, []);

  // Don't render anything if ads are disabled
  if (shouldDisableAds()) {
    return null;
  }

  if (isLoading) {
    return (
      <Skeleton
        loading
        className={ className }
        h={{ base: 12, lg: 6 }}
        w="100%"
        flexGrow={ 1 }
        maxW="800px"
        display="block"
      />
    );
  }

  if (!adData) {
    return null;
  }

  const urlObject = new URL(adData.ad.url);

  return (
    <Box className={ className }>
      <Text
        as="span"
        whiteSpace="pre-wrap"
        fontWeight={ 700 }
        mr={ 3 }
        display={{ base: 'none', lg: 'inline' }}
      >
        Ads:
      </Text>
      { urlObject.hostname === 'nifty.ink' ?
        <Text as="span" mr={ 1 }>🎨</Text> : (
          <Image
            src={ adData.ad.thumbnail }
            width="20px"
            height="20px"
            verticalAlign="text-bottom"
            mr={ 1 }
            display="inline-block"
            alt=""
          />
        ) }
      <Text as="span" whiteSpace="pre-wrap">{ `${ adData.ad.name } ${ ndash } ${ adData.ad.description_short } ` }</Text>
      <Link href={ adData.ad.url } external noIcon>{ adData.ad.cta_button }</Link>
    </Box>
  );
};

export default chakra(CoinzillaTextAd);
