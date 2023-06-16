// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import ResourcesTable from '../../features/resources/ResourcesTable';
import { MetricsContextProvider } from '../../features/context/MetricsContext';

function ResourcesPage() {
  return (
    <MetricsContextProvider>
      <ResourcesTable/>
    </MetricsContextProvider>
  );
}

export default ResourcesPage;
