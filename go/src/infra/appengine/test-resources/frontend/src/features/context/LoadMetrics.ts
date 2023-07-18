// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { FetchDirectoryMetricsRequest, fetchDirectoryMetrics, DirectoryNode,
  FetchTestMetricsRequest,
  FetchTestMetricsResponse,
  MetricType,
  Period,
  TestDateMetricData,
  MetricsDateMap,
  fetchTestMetrics,
  FetchDirectoryMetricsResponse } from '../../api/resources';
import { formatDate } from '../../utils/formatUtils';
import { Node, Params, Path } from './MetricsContext';

export function computeDates(params: Params): string[] {
  const computedDates: string[] = [];
  const datesBefore = params.timelineView ? 4 : 0;
  for (let x = datesBefore; x >= 0; x--) {
    const newDate = new Date(params.date);
    newDate.setDate(
        params.date.getDate() - (x * (params.period === Period.DAY ? 1 : 7)),
    );
    computedDates.push(formatDate(newDate));
  }
  return computedDates;
}

export function loadTestMetrics(
    component: string,
    params: Params,
    successCallback: (response: FetchTestMetricsResponse, fetchedDates: string[]) => void,
    failureCallback: (erorr: any) => void,
) {
  const datesToFetch = computeDates(params);
  const request: FetchTestMetricsRequest = {
    components: [component],
    period: params.period,
    dates: datesToFetch,
    metrics: [
      MetricType.NUM_RUNS,
      MetricType.AVG_RUNTIME,
      MetricType.TOTAL_RUNTIME,
      MetricType.NUM_FAILURES,
      MetricType.AVG_CORES,
    ],
    filter: params.filter,
    page_offset: params.page * params.rowsPerPage,
    page_size: params.rowsPerPage,
    sort: {
      metric: params.sort,
      ascending: params.ascending,
    },
  };
  fetchTestMetrics(request).then((response) => {
    successCallback(response, datesToFetch);
  }).catch(failureCallback);
}

type DataAction =
 | { type: 'merge_test', tests: TestDateMetricData[] }
 | {
  type: 'merge_dir',
  nodes: DirectoryNode[],
  onExpand: (node: Node) => void,
  parentId?: string
}

function findNode(nodes: Node[], id: string): Node | undefined {
  for (let i = 0; i < nodes.length; i++) {
    if (nodes[i].id === id) {
      return nodes[i];
    } else if (nodes[i].nodes.length > 0) {
      return findNode(nodes[i].nodes, id);
    }
  }
  return undefined;
}

export function dataReducer(state: Node[], action: DataAction): Node[] {
  switch (action.type) {
    case 'merge_test':
      if (action.tests === undefined) {
        return [];
      }
      return action.tests.map((test) => ({
        id: test.testId,
        name: test.testName,
        fileName: test.fileName,
        metrics: createMetricsMap(test.metrics),
        isLeaf: false,
        nodes: test.variants.map((variant) => ({
          id: `${test.testId}:${variant.builder}:${variant.suite}`,
          name: variant.builder,
          subname: variant.suite,
          metrics: createMetricsMap(variant.metrics),
          isLeaf: true,
          nodes: [],
        })),
      }));
    case 'merge_dir': {
      const nodes = action.nodes ? action.nodes.map((node) => ({
        id: node.id,
        path: node.id,
        name: node.name,
        metrics: createMetricsMap(node.metrics),
        isLeaf: false,
        onExpand: action.onExpand,
        loaded: false,
        nodes: [],
      })) : [];
      if (action.parentId === undefined) {
        return nodes;
      } else {
        const parentNode = findNode(state, action.parentId);
        if (parentNode !== undefined) {
          parentNode.nodes = nodes;
          (parentNode as Path).loaded = true;
        }
        // Necessary to return a new object to trigger a re-render.
        return [...state];
      }
    }
  }
  return state;
}

export function loadDirectoryMetrics(
    component: string,
    params: Params,
    parentId: string,
    successCallback: (response: FetchDirectoryMetricsResponse, fetchedDates: string[]) => void,
    failureCallback: (erorr: any) => void,
) {
  const datesToFetch = computeDates(params);
  const request: FetchDirectoryMetricsRequest = {
    components: [component],
    period: params.period,
    dates: datesToFetch,
    parent_ids: [parentId],
    metrics: [
      MetricType.NUM_RUNS,
      MetricType.AVG_RUNTIME,
      MetricType.TOTAL_RUNTIME,
      MetricType.NUM_FAILURES,
      MetricType.AVG_CORES,
    ],
    filter: params.filter,
    sort: {
      metric: params.sort,
      ascending: params.ascending,
    },
  };
  fetchDirectoryMetrics(request).then((response) => {
    successCallback(response, datesToFetch);
  }).catch(failureCallback);
}

function createMetricsMap(
    metrics: MetricsDateMap,
): Map<string, Map<MetricType, number>> {
  const ret = new Map<string, Map<MetricType, number>>();
  for (const date in metrics) {
    if (Object.hasOwn(metrics, date)) {
      const metricMap = new Map<MetricType, number>();
      metrics[date].data.forEach((metric) => {
        metricMap.set(metric.metricType, metric.metricValue);
      });
      ret.set(date, metricMap);
    }
  }
  return ret;
}

