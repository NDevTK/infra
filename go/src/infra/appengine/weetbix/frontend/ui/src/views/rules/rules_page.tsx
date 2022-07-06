// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { useParams, Link } from 'react-router-dom';

import Button from '@mui/material/Button';
import Grid from '@mui/material/Grid';
import Container from '@mui/material/Container';
import HelpTooltip from '../../components/help_tooltip/help_tooltip';
import RulesTable from '../../components/rules_table/rules_table';

const rulesDescription = 'Rules define an association between failures and bugs. Weetbix uses these ' +
  'associations to calculate bug impact, automatically adjust bug priority and verified status, and ' +
  'to surface bugs for failures in the MILO test results UI.';

const RulesPage = () => {
  const { project } = useParams();
  return (
    <Container maxWidth={false}>
      <Grid container>
        <Grid item xs={8}>
          <h2>Rules in project {project}<HelpTooltip text={rulesDescription}></HelpTooltip></h2>
        </Grid>
        <Grid item xs={4} sx={{ textAlign: 'right' }}>
          <Button component={Link} variant='contained' to='new' sx={{ marginBlockStart: '20px' }}>New Rule</Button>
        </Grid>
      </Grid>
      {(project) && (
        <RulesTable project={project}></RulesTable>
      )}
    </Container>
  );
};

export default RulesPage;

