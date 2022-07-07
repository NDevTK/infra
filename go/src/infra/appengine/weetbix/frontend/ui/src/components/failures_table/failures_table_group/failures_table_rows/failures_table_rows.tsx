// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import dayjs from 'dayjs';
import { ReactNode, useState } from 'react';

import ArrowDropDownIcon from '@mui/icons-material/ArrowDropDown';
import ArrowRightIcon from '@mui/icons-material/ArrowRight';
import Grid from '@mui/material/Grid';
import IconButton from '@mui/material/IconButton';
import Link from '@mui/material/Link';
import TableCell from '@mui/material/TableCell';
import TableRow from '@mui/material/TableRow';

import {
  ClusterFailure,
  FailureGroup,
  VariantGroup,
} from '../../../../tools/failures_tools';
import { failureLink } from '../../../../tools/urlHandling/links';

interface Props {
  group: FailureGroup;
  variantGroups: VariantGroup[];
  children?: ReactNode;
}

const FailuresTableRows = ({
  group,
  variantGroups,
  children = null,
}: Props) => {
  const [expanded, setExpanded] = useState(false);

  const toggleExpand = () => {
    setExpanded(!expanded);
  };

  const ungroupedVariants = (failure: ClusterFailure) => {
    const unselectedVariants = variantGroups
        .filter((v) => !v.isSelected)
        .map((v) => v.key);
    return unselectedVariants
        .map((key) => failure.variant?.filter((v) => v.key == key)?.[0])
        .filter((v) => v);
  };

  return (
    <>
      <TableRow>
        <TableCell
          key={group.id}
          sx={{
            paddingLeft: `${20 * group.level}px`,
            width: '60%',
          }}
          data-testid="failures_table_group_cell"
        >
          {group.failure ? (
            <>
              <Link
                aria-label="Failure invocation id"
                sx={{ mr: 2 }}
                href={failureLink(group.failure)}
                target="_blank"
              >
                {group.failure.ingestedInvocationId}
              </Link>
              <small data-testid="ungrouped_variants">
                {ungroupedVariants(group.failure)
                    .map((v) => v && `${v.key}: ${v.value}`)
                    .filter(v => v)
                    .join(', ')}
              </small>
            </>
          ) : (
            <Grid
              container
              justifyContent="start"
              alignItems="baseline"
              columnGap={2}
              flexWrap="nowrap"
            >
              <Grid item>
                <IconButton
                  aria-label="Expand group"
                  onClick={() => toggleExpand()}
                >
                  {expanded ? <ArrowDropDownIcon /> : <ArrowRightIcon />}
                </IconButton>
              </Grid>
              <Grid item sx={{ overflowWrap: 'anywhere' }}>{group.name || 'none'}</Grid>
            </Grid>
          )}
        </TableCell>
        <TableCell data-testid="failure_table_group_presubmitrejects">
          {group.failure ? (
            <>
              {group.failure.presubmitRunId ? (
                <Link
                  aria-label="Presubmit rejects link"
                  href={`https://luci-change-verifier.appspot.com/ui/run/${group.failure.presubmitRunId.id}`}
                  target="_blank"
                >
                  {group.presubmitRejects}
                </Link>
              ) : (
                '-'
              )}
            </>
          ) : (
            group.presubmitRejects
          )}
        </TableCell>
        <TableCell className="number">{group.invocationFailures}</TableCell>
        <TableCell className="number">{group.criticalFailuresExonerated}</TableCell>
        <TableCell className="number">{group.failures}</TableCell>
        <TableCell>{dayjs(group.latestFailureTime).fromNow()}</TableCell>
      </TableRow>
      {/** Render the remaining rows in the group */}
      {expanded && children}
    </>
  );
};

export default FailuresTableRows;
