// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { Paper } from '@mui/material';
import ResourcesTable from '../../features/resources/ResourcesTable';
import { MetricsContextProvider } from '../../features/context/MetricsContext';
import ResourcesToolbar from '../../features/resources/ResourcesToolbar';
import ResourcesSearchParams from '../../features/resources/ResourcesSearchParams';
import ComponentParams from '../../features/components/ComponentParams';

function ResourcesPage() {
  return (
    <MetricsContextProvider>
      <ResourcesToolbar/>
      <Paper sx={{ margin: '10px 20px' }}>
        <ResourcesTable/>
      </Paper>
      <ResourcesSearchParams/>
      <ComponentParams/>
    </MetricsContextProvider>
  );
}

export default ResourcesPage;
