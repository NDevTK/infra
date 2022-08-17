// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import './styles/style.css';

import * as React from 'react';
import { QueryClient, QueryClientProvider } from 'react-query';
import { Route, Routes } from 'react-router-dom';

import { BaseLayout } from './src/layouts/base';
import { AnalysisDetailsPage } from './src/views/analysis_details/analysis_details';
import { FailureAnalysesPage } from './src/views/failure_analyses';
import { StatisticsPage } from './src/views/statistics';
import { TriggerAnalysisPage } from './src/views/trigger_analysis';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
    },
  },
});

export const App = () => {
  return (
    <QueryClientProvider client={queryClient}>
      <Routes>
        <Route path='/' element={<BaseLayout />}>
          <Route index element={<FailureAnalysesPage />} />
          <Route path='trigger' element={<TriggerAnalysisPage />} />
          <Route path='analysis/b/:buildID' element={<AnalysisDetailsPage />} />
          <Route path='statistics' element={<StatisticsPage />} />
        </Route>
      </Routes>
    </QueryClientProvider>
  );
};
