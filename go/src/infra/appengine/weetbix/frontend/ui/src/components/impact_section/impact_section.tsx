// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { useQuery } from 'react-query';
import { useParams } from 'react-router-dom';

import Container from '@mui/material/Container';
import CircularProgress from '@mui/material/CircularProgress';
import Grid from '@mui/material/Grid';
import Paper from '@mui/material/Paper';

import { getClustersService, BatchGetClustersRequest } from '../../services/cluster';
import ErrorAlert from '../error_alert/error_alert';
import ImpactTable from '../impact_table/impact_table';

const ImpactSection = () => {
  const { project, algorithm, id } = useParams();
  let currentAlgorithm = algorithm;
  if (!currentAlgorithm) {
    currentAlgorithm = 'rules';
  }
  const clustersService = getClustersService();
  const { isLoading, isError, isSuccess, data: cluster, error } = useQuery(['cluster', `${currentAlgorithm}:${id}`], async () => {
    const request: BatchGetClustersRequest = {
      parent: `projects/${encodeURIComponent(project || '')}`,
      names: [
        `projects/${encodeURIComponent(project || '')}/clusters/${encodeURIComponent(currentAlgorithm || '')}/${encodeURIComponent(id || '')}`,
      ],
    };

    const response = await clustersService.batchGet(request);

    // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
    return response.clusters![0];
  });

  return (
    <Paper elevation={3} sx={{ pt: 1, pb: 4 }}>
      <Container maxWidth={false}>
        <h2>Impact</h2>
        {
          isLoading && (
            <Grid container item alignItems="center" justifyContent="center">
              <CircularProgress />
            </Grid>
          )
        }
        {
          isError && (
            <ErrorAlert
              errorText={`Got an error while loading the cluster: ${error}`}
              errorTitle="Failed to load cluster"
              showError/>
          )
        }
        {
          isSuccess && cluster && (
            <ImpactTable cluster={cluster}></ImpactTable>
          )
        }
      </Container>
    </Paper>
  );
};

export default ImpactSection;
