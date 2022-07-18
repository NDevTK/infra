// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import dayjs from 'dayjs';

import { useQuery } from 'react-query';

import Box from '@mui/material/Box';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import LinearProgress from '@mui/material/LinearProgress';
import Link from '@mui/material/Link';

import { getRulesService, ListRulesRequest } from '../../services/rules';
import { linkToRule } from '../../tools/urlHandling/links';
import ErrorAlert from '../error_alert/error_alert';

interface Props {
  project: string;
}

const RulesTable = ({ project } : Props ) => {
  const rulesService = getRulesService();
  const { isLoading, isError, data: rules, error } = useQuery(['rules', project], async () => {
    const request: ListRulesRequest = {
      parent: `projects/${encodeURIComponent(project || '')}`,
    };

    const response = await rulesService.list(request);

    const sortedRules = response.rules.sort((a, b)=> {
      // These are RFC 3339-formatted date/time strings.
      // Because they are all use the same timezone, and RFC 3339
      // date/times are specified from most significant to least
      // significant, any string sort that produces a lexicographical
      // ordering should also sort by time.
      return b.lastUpdateTime.localeCompare(a.lastUpdateTime);
    });
    return sortedRules;
  });
  if (isLoading) {
    return <LinearProgress />;
  }

  if (isError || rules === undefined) {
    return <ErrorAlert
      errorText={`Got an error while loading rules: ${error}`}
      errorTitle="Failed to load rules"
      showError/>;
  }

  return (
    <TableContainer component={Box}>
      <Table data-testid="impact-table" size="small" sx={{ overflowWrap: 'anywhere' }}>
        <TableHead>
          <TableRow>
            <TableCell>Rule Definition</TableCell>
            <TableCell sx={{ width: '130px' }}>Bug</TableCell>
            <TableCell sx={{ width: '100px' }}>Last Updated</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {
            rules.map((rule) => (
              <TableRow key={rule.ruleId}>
                <TableCell><Link href={linkToRule(rule.project, rule.ruleId)} underline="hover">{rule.ruleDefinition}</Link></TableCell>
                <TableCell><Link href={rule.bug.url} underline="hover">{rule.bug.linkText}</Link></TableCell>
                <TableCell>{dayjs.utc(rule.lastUpdateTime).local().fromNow()}</TableCell>
              </TableRow>
            ))
          }
        </TableBody>
      </Table>
    </TableContainer>
  );
};

export default RulesTable;
