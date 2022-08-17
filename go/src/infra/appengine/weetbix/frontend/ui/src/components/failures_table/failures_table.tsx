// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {
  useEffect,
  useState,
} from 'react';
import { useQuery } from 'react-query';
import {
  useUpdateEffect,
} from 'react-use';

import CircularProgress from '@mui/material/CircularProgress';
import Grid from '@mui/material/Grid';
import { SelectChangeEvent } from '@mui/material/Select';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';

import { getClustersService } from '../../services/cluster';
import {
  countAndSortFailures,
  countDistictVariantValues,
  defaultFailureFilter,
  defaultImpactFilter,
  FailureFilter,
  FailureFilters,
  FailureGroup,
  VariantGroup,
  groupAndCountFailures,
  ImpactFilter,
  ImpactFilters,
  MetricName,
  sortFailureGroups,
} from '../../tools/failures_tools';
import ErrorAlert from '../error_alert/error_alert';
import FailuresTableFilter from './failures_table_filter/failures_table_filter';
import FailuresTableGroup from './failures_table_group/failures_table_group';
import FailuresTableHead from './failures_table_head/failures_table_head';

interface Props {
    project: string;
    clusterAlgorithm: string;
    clusterId: string;
}

const FailuresTable = ({
  project,
  clusterAlgorithm,
  clusterId,
}: Props) => {
  const [groups, setGroups] = useState<FailureGroup[]>([]);
  const [variantGroups, setVariantGroups] = useState<VariantGroup[]>([]);

  const [failureFilter, setFailureFilter] = useState<FailureFilter>(defaultFailureFilter);
  const [impactFilter, setImpactFilter] = useState<ImpactFilter>(defaultImpactFilter);
  const [selectedVariantGroups, setSelectedVariantGroups] = useState<string[]>([]);

  const [sortMetric, setCurrentSortMetric] = useState<MetricName>('latestFailureTime');
  const [isAscending, setIsAscending] = useState(false);

  const {
    isLoading,
    isError,
    data: failures,
    error,
  } = useQuery(
      ['clusterFailures', project, clusterAlgorithm, clusterId],
      async () => {
        const service = getClustersService();
        const response = await service.queryClusterFailures({
          parent: `projects/${project}/clusters/${clusterAlgorithm}/${clusterId}/failures`,
        });
        return response.failures || [];
      });

  useEffect( () => {
    if (failures) {
      setVariantGroups(countDistictVariantValues(failures));
    }
  }, [failures]);

  useUpdateEffect(() => {
    setGroups(sortFailureGroups(groups, sortMetric, isAscending));
  }, [sortMetric, isAscending]);

  useUpdateEffect(() => {
    setGroups(countAndSortFailures(groups, impactFilter));
  }, [impactFilter]);

  useUpdateEffect(() => {
    groupCountAndSortFailures();
  }, [failureFilter]);

  useUpdateEffect(() => {
    groupCountAndSortFailures();
  }, [variantGroups]);

  useUpdateEffect(() => {
    const variantGroupsClone = [...variantGroups];
    variantGroupsClone.forEach((variantGroup) => {
      variantGroup.isSelected = selectedVariantGroups.includes(variantGroup.key);
    });
    setVariantGroups(variantGroupsClone);
  }, [selectedVariantGroups]);

  const groupCountAndSortFailures = () => {
    if (failures) {
      let updatedGroups = groupAndCountFailures(failures, variantGroups, failureFilter);
      updatedGroups = countAndSortFailures(updatedGroups, impactFilter);
      setGroups(sortFailureGroups(updatedGroups, sortMetric, isAscending));
    }
  };

  const onImpactFilterChanged = (event: SelectChangeEvent) => {
    setImpactFilter(ImpactFilters.filter((filter) => filter.name === event.target.value)?.[0] || ImpactFilters[1]);
  };

  const onFailureFilterChanged = (event: SelectChangeEvent) => {
    setFailureFilter((event.target.value as FailureFilter) || FailureFilters[0]);
  };

  const handleVariantsChange = (event: SelectChangeEvent<typeof selectedVariantGroups>) => {
    const value = event.target.value;
    setSelectedVariantGroups(typeof value === 'string' ? value.split(',') : value);
  };

  const toggleSort = (metric: MetricName) => {
    if (metric === sortMetric) {
      setIsAscending(!isAscending);
    } else {
      setCurrentSortMetric(metric);
      setIsAscending(false);
    }
  };

  if (isLoading) {
    return (
      <Grid container item alignItems="center" justifyContent="center">
        <CircularProgress />
      </Grid>
    );
  }

  if (isError || !failures) {
    return (
      <ErrorAlert
        errorTitle="Failed to load failures"
        errorText={`Loading cluster failures failed due to: ${error}`}
        showError/>
    );
  }

  return (
    <Grid container columnGap={2} rowGap={2}>
      <FailuresTableFilter
        failureFilter={failureFilter}
        onFailureFilterChanged={onFailureFilterChanged}
        impactFilter={impactFilter}
        onImpactFilterChanged={onImpactFilterChanged}
        variantGroups={variantGroups}
        selectedVariantGroups={selectedVariantGroups}
        handleVariantGroupsChange={handleVariantsChange}/>
      <Grid item xs={12}>
        <Table size="small">
          <FailuresTableHead
            toggleSort={toggleSort}
            sortMetric={sortMetric}
            isAscending={isAscending}/>
          <TableBody>
            {
              groups.map((group) => (
                <FailuresTableGroup
                  project={project}
                  parentKeys={[]}
                  key={group.id}
                  group={group}
                  variantGroups={variantGroups}/>
              ))
            }
          </TableBody>
        </Table>
      </Grid>
    </Grid>
  );
};

export default FailuresTable;
