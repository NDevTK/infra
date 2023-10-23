// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { Auth } from '../../../api/auth';
import { FetchDirectoryMetricsRequest, fetchDirectoryMetrics, DirectoryNode,
  FetchTestMetricsRequest,
  FetchTestMetricsResponse,
  MetricType,
  Period,
  TestDateMetricData,
  MetricsDateMap,
  fetchTestMetrics,
  FetchDirectoryMetricsResponse,
  DirectoryNodeType,
} from '../../../api/resources';
import { formatDate } from '../../../utils/formatUtils';
import { Node, Params, Path, Test, isPath } from './TestMetricsContext';

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
    auth: Auth,
    components: string[],
    params: Params,
    successCallback: (
      response: FetchTestMetricsResponse,
      fetchedDates: string[]
      ) => void,
    failureCallback: (erorr: any) => void,
    fileNames?: string[],
) {
  const datesToFetch = computeDates(params);
  const request: FetchTestMetricsRequest = {
    components: components,
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
      sort_date: datesToFetch[params.sortIndex],
    },
  };
  if (fileNames) {
    request.file_names = fileNames;
  }
  fetchTestMetrics(auth, request).then((response) => {
    successCallback(response, datesToFetch);
  }).catch(failureCallback);
}

type DataAction =
 | {
  type: 'merge_test',
  tests: TestDateMetricData[],
  parentId?: string,
  footer?: JSX.Element,
 }
 | {
  type: 'merge_dir',
  nodes: DirectoryNode[],
  onExpand: (node: Node) => void,
  parentId?: string
 }
 | {
  type: 'rebuild_state',
  nodes: DirectoryNode[],
  tests: TestDateMetricData[],
  onExpand?: (node: Node) => void,
 }

function findNode(nodes: Node[], id: string): Node | undefined {
  for (let i = 0; i < nodes.length; i++) {
    if (nodes[i].id === id) {
      return nodes[i];
    } else if (nodes[i].rows.length > 0) {
      const node = findNode(nodes[i].rows, id);
      if (node !== undefined) {
        return node;
      }
    }
  }
  return undefined;
}

export function getLoadedParentIds(
    state: Node[],
    dirs: string[] = [],
    files: string[] = [],
): [dirs: string[], files: string[]] {
  state.forEach((node) => {
    if (isPath(node) && node.loaded) {
      if (node.type === DirectoryNodeType.FILENAME) {
        files.push(node.id);
      } else {
        dirs.push(node.id);
      }
    }
    if (node.rows.length > 0) {
      getLoadedParentIds(node.rows, dirs, files);
    }
  });
  return [dirs, files];
}

function createTestNode(test: TestDateMetricData): Test {
  return {
    id: test.testId,
    name: test.testName,
    fileName: test.fileName,
    metrics: createMetricsMap(test.metrics),
    isExpandable: true,
    rows: test.variants.map((variant) => ({
      id: `${test.testId}:${variant.bucket}:${variant.builder}` +
        `:${variant.suite}`,
      name: variant.bucket + '/' + variant.builder,
      subname: variant.suite,
      metrics: createMetricsMap(variant.metrics),
      isExpandable: false,
      rows: [],
    })),
  };
}

function createPathNode(
    node: DirectoryNode,
    onExpand?: (node: Node) => void,
) : Path {
  return {
    id: node.id,
    path: node.id,
    name: node.name + ((node.type === DirectoryNodeType.DIRECTORY) ? '/' : ''),
    metrics: createMetricsMap(node.metrics),
    isExpandable: true,
    onExpand: onExpand,
    loaded: false,
    type: node.type as DirectoryNodeType,
    rows: [],
  };
}

export function dataReducer(state: Node[], action: DataAction): Node[] {
  let nodes: Node[] = [];
  switch (action.type) {
    case 'merge_test': {
      nodes = action.tests?.map(createTestNode) || [];
      break;
    }
    case 'merge_dir': {
      nodes = action.nodes?.map(
          (node) => {
            return createPathNode(node, action.onExpand);
          },
      ) || [];
      break;
    }
    case 'rebuild_state': {
      return rebuildState(action.nodes, action.tests, action.onExpand);
    }
  }
  if (action.parentId === undefined) {
    return nodes;
  } else {
    const parentNode = findNode(state, action.parentId);
    if (parentNode !== undefined) {
      parentNode.rows = nodes;
      (parentNode as Path).loaded = true;
      if (action.type === 'merge_test') {
        parentNode.footer = action.footer;
      }
    }
    // Necessary to return a new object to trigger a re-render.
    return [...state];
  }
}

function getParentId(id: string) {
  const pieces = id.split('/');
  pieces.pop();
  // Just to note that in javascript, ['', ''] joined becomes '/'
  // So parent ID for '//foo' is just '/' and not '//'
  return pieces.join('/');
}

// This function takes a list of paths and a list of tests and rebuilds the
// directory tree. The basic algorithm is that it iterates through the paths
// and tests, figures out the parent ID of each node, and populates a map of
// parent IDs to child nodes. It then starts from the root nodes and sets
// the children for each node it encounters based on the populated map.
function rebuildState(
    paths: DirectoryNode[],
    tests: TestDateMetricData[],
    onExpand?: (node: Node) => void,
): Node[] {
  const parents = new Map<string, Node[]>();
  paths.forEach((path) => {
    const node = createPathNode(path, onExpand);
    const parentId = getParentId(node.path);
    if (!parents.has(parentId)) {
      parents.set(parentId, []);
    }
    parents.get(parentId)?.push(node);
  });
  tests.forEach((test) => {
    const node = createTestNode(test);
    if (!parents.has(node.fileName)) {
      parents.set(node.fileName, []);
    }
    parents.get(node.fileName)?.push(node);
  });
  const nodes = parents.get('/') || [];
  const populate = (nodes: Node[]) => {
    nodes.forEach((node) => {
      if (isPath(node) && parents.has(node.id)) {
        node.rows = parents.get(node.id) || [];
        node.loaded = true;
        populate(node.rows);
      }
    });
  };
  populate(nodes);
  return nodes;
}

export function loadDirectoryMetrics(
    auth: Auth,
    components: string[],
    params: Params,
    parentIds: string[],
    successCallback: (
      response: FetchDirectoryMetricsResponse,
      fetchedDates: string[]
      ) => void,
    failureCallback: (erorr: any) => void,
) {
  const datesToFetch = computeDates(params);
  const request: FetchDirectoryMetricsRequest = {
    components: components,
    period: params.period,
    dates: datesToFetch,
    parent_ids: parentIds,
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
      sort_date: datesToFetch[params.sortIndex],
    },
  };
  fetchDirectoryMetrics(auth, request).then((response) => {
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

