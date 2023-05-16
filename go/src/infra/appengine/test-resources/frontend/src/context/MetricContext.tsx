// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import React, { createContext, useState } from 'react';
import { TestDateMetricData } from '@/api/resources';

type MetricContextProviderProps = {
  children: React.ReactNode
}

interface MetricContextValue {
  testDateMetricData: TestDateMetricData[] | null;
  setMetrics: React.Dispatch<React.SetStateAction<TestDateMetricData[] | null>>;
}

const MetricContext = createContext<MetricContextValue
| null>(null);

export const MetricContextProvider = ({ children } : MetricContextProviderProps) => {
  const [testDateMetricData, setMetrics] = useState<TestDateMetricData[] | null>(null);
  return (
    <MetricContext.Provider value={{ testDateMetricData, setMetrics }}>
      { children }
    </MetricContext.Provider>
  );
};

export default MetricContext;
