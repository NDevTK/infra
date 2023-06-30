// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import ResourcesTable from '../../features/resources/ResourcesTable';
import { MetricsContextProvider } from '../../features/context/MetricsContext';
import ResourcesToolbar from '../../features/resources/ResourcesToolbar';
import ResourcesSearchParams from '../../features/resources/ResourcesSearchParams';

function ResourcesPage() {
  return (
    <MetricsContextProvider>
      <ResourcesToolbar/>
      <ResourcesTable/>
      <ResourcesSearchParams/>
    </MetricsContextProvider>
  );
}

export default ResourcesPage;
