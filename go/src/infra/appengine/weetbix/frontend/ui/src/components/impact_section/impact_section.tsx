// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { useQuery } from 'react-query';
import { useParams } from 'react-router-dom';

import Container from '@mui/material/Container';
import LinearProgress from '@mui/material/LinearProgress';
import Paper from '@mui/material/Paper';

import { getCluster } from '../../services/cluster';
import ErrorAlert from '../error_alert/error_alert';
import FailuresTable from '../failures_table/failures_table';
import ImpactTable from '../impact_table/impact_table';

const ImpactSection = () => {
  const { project, algorithm, id } = useParams();
  let currentAlgorithm = algorithm;
  if (!currentAlgorithm) {
    currentAlgorithm = 'rules-v2';
  }
  const { isLoading, isError, data: cluster, error } = useQuery(['cluster', `${currentAlgorithm}:${id}`], () => {
    // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
    return getCluster(project!, currentAlgorithm!, id!);
  });

  if (isLoading) {
    return <LinearProgress />;
  }

  if (isError || !cluster) {
    return <ErrorAlert
      errorText={`Got an error while loading the cluster: ${error}`}
      errorTitle="Failed to load cluster"
      showError/>;
  }

  return (
    <Paper elevation={3} sx={{ pt: 1, pb: 4 }}>
      <Container maxWidth={false}>
        <h2>Impact</h2>
        <ImpactTable cluster={cluster}></ImpactTable>
        <h2>Recent Failures</h2>
        {
          (id && project) && (
            <FailuresTable
              clusterAlgorithm={currentAlgorithm}
              clusterId={id}
              project={project}/>
          )
        }
      </Container>
    </Paper>
  );
};

export default ImpactSection;
