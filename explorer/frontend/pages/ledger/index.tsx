import React from 'react';

import type { NextPageWithLayout } from 'nextjs/types';

import PageNextJs from 'nextjs/PageNextJs';

import LedgerPage from 'ui/pages/Ledger';
import LayoutApp from 'ui/shared/layout/LayoutApp';

const Page: NextPageWithLayout = () => {
  return (
    <PageNextJs pathname="/ledger">
      <LedgerPage/>
    </PageNextJs>
  );
};

Page.getLayout = function getLayout(page: React.ReactElement) {
  return (
    <LayoutApp>
      { page }
    </LayoutApp>
  );
};

export default Page;

export { base as getServerSideProps } from 'nextjs/getServerSideProps'; 