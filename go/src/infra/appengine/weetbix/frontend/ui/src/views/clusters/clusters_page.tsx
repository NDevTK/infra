// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { useParams } from 'react-router-dom';

import Grid from '@mui/material/Grid';
import Container from '@mui/material/Container';
import HelpTooltip from '../../components/help_tooltip/help_tooltip';
import ClustersTable from '../../components/clusters_table/clusters_table';

const rulesDescription = 'Clusters are groups of related test failures. Weetbix\'s clusters ' +
  'comprise clusters identified by algorithms (based on test name or failure reason) ' +
  'and clusters defined by a failure association rule (where the cluster contains all failures ' +
  'associated with a specific bug).';

const ClustersPage = () => {
  const { project } = useParams();
  return (
    <Container maxWidth={false}>
      <Grid container>
        <Grid item xs={8}>
          <h2>Clusters in project {project}<HelpTooltip text={rulesDescription}></HelpTooltip></h2>
        </Grid>
      </Grid>
      {(project) && (
        <ClustersTable project={project}></ClustersTable>
      )}
    </Container>
  );
};

export default ClustersPage;

