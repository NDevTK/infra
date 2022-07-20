// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { useParams } from 'react-router-dom';

import Container from '@mui/material/Container';
import Paper from '@mui/material/Paper';

import FailuresTable from '../failures_table/failures_table';

const RecentFailuresSection = () => {
  const { project, algorithm, id } = useParams();
  let currentAlgorithm = algorithm;
  if (!currentAlgorithm) {
    currentAlgorithm = 'rules-v2';
  }

  return (
    <Paper elevation={3} sx={{ pt: 1, pb: 4 }}>
      <Container maxWidth={false}>
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

export default RecentFailuresSection;
