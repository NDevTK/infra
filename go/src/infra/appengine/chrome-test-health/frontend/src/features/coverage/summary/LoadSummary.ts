// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {
  CoverageMetric,
  GetProjectDefaultConfigRequest,
  GetProjectDefaultConfigResponse,
  GetSummaryByComponentRequest,
  GetSummaryCoverageRequest,
  Platform,
  SummaryNode,
  getProjectDefaultConfig,
  getSummaryCoverage,
  getSummaryCoverageByComponent,
} from '../../../api/coverage';
import { Auth } from '../../../api/auth';
import { Row } from '../../../components/table/DataTable';

export enum DirectoryNodeType {
  DIRECTORY = 'DIRECTORY',
  FILENAME = 'FILENAME',
}

export enum MetricType {
  LINE = 'LINE'
}

export interface MetricData {
  covered: number,
  total: number,
  percentageCovered: number
}

export interface Node extends Row<Node> {
  name: string,
  metrics: Map<MetricType, MetricData>,
  rows: Node[],
}

export interface Path extends Node {
  path: string,
  type: DirectoryNodeType,
  loaded: boolean,
}

export interface Params {
  host: string,
  project: string,
  gitilesRef: string,
  revision: string,
  unitTestsOnly: boolean,
  platform: string,
  builder: string,
  bucket: string,
  platformList: Platform[]
}

export enum DataActionType {
  MERGE_DIR = 'merge_dir',
  BUILD_TREE = 'build_tree',
  CLEAR_DIR = 'clear_dir'
}

type DataAction =
  | {
    type: DataActionType.MERGE_DIR,
    summaryNodes: SummaryNode[],
    loaded: boolean,
    onExpand: (node: Node) => void,
    parentId?: string
  }
  | {
    type: DataActionType.BUILD_TREE,
    summaryNodes: SummaryNode[],
    onExpand: (node: Node) => void,
  }
  | {
    type: DataActionType.CLEAR_DIR
  }

export function dataReducer(state: Node[], action: DataAction): Node[] {
  let nodes: Node[] = [];
  switch (action.type) {
    case DataActionType.MERGE_DIR: {
      nodes = action.summaryNodes.map(
          (node) => {
            return createPathNode(
                node,
            (node.isDir) ?
              DirectoryNodeType.DIRECTORY :
              DirectoryNodeType.FILENAME,
            action.loaded,
            action.onExpand,
            );
          },
      );
      if (action.parentId === undefined) {
        return nodes;
      } else {
        const parentNode = findNode(state, action.parentId);
        if (parentNode !== undefined) {
          parentNode.rows = nodes;
          (parentNode as Path).loaded = true;
        }
        return [...state];
      }
    }
    case DataActionType.BUILD_TREE: {
      action.summaryNodes.forEach((summaryNode) => {
        nodes.push(mergeSubTree(summaryNode, action.onExpand));
      });
      return nodes;
    }
    case DataActionType.CLEAR_DIR: {
      return [] as Node[];
    }
  }
}

export function loadProjectDefaultConfig(
    auth: Auth,
    LuciProject: string,
    revision: string,
    ModifierId: string,
    successCallback: (
    response: GetProjectDefaultConfigResponse
  ) => void,
    failureCallback: (error: any) => void,
) {
  const request: GetProjectDefaultConfigRequest = {
    luci_project: LuciProject,
    revision: revision,
    modifier_id: ModifierId,
  };

  getProjectDefaultConfig(auth, request).then((response) => {
    successCallback(response);
  }).catch(failureCallback);
}

export function loadSummary(
    auth: Auth,
    params: Params,
    path: string,
    successCallback: (
    response: SummaryNode[],
  ) => void,
    failureCallback: (error: any) => void,
) {
  const request: GetSummaryCoverageRequest = {
    gitiles_host: params.host,
    gitiles_project: params.project,
    gitiles_ref: params.gitilesRef,
    gitiles_revision: params.revision,
    path: path,
    unit_tests_only: params.unitTestsOnly,
    data_type: 'dirs',
    builder: params.builder,
    bucket: params.bucket,
  };

  getSummaryCoverage(auth, request).then((response) => {
    successCallback(response);
  }).catch(failureCallback);
}

export function loadSummaryByComponents(
    auth: Auth,
    params: Params,
    components: string[],
    successCallback: (
    response: SummaryNode[]
  ) => void,
    failureCallback: (error: any) => void,
) {
  const request: GetSummaryByComponentRequest = {
    gitiles_host: params.host,
    gitiles_project: params.project,
    gitiles_ref: params.gitilesRef,
    gitiles_revision: params.revision,
    components,
    unit_tests_only: params.unitTestsOnly,
    builder: params.builder,
    bucket: params.bucket,
  };

  getSummaryCoverageByComponent(auth, request).then((response) => {
    successCallback(response);
  }).catch(failureCallback);
}

// -------------- HELPER FUNCTIONS --------------

function createPathNode(
    summaryNode: SummaryNode,
    nodeType: DirectoryNodeType,
    loaded: boolean,
    onExpand?: (node: Node) => void,
): Path {
  return {
    id: summaryNode.name,
    path: summaryNode.path,
    name: summaryNode.name,
    metrics: createCoverageMap(summaryNode.summaries),
    isExpandable: summaryNode.isDir,
    onExpand: onExpand,
    loaded,
    type: nodeType,
    rows: [],
  };
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

function mergeSubTree(
    summaryNode: SummaryNode,
    onExpand: (node: Node) => void,
): Node {
  const root: Node = createPathNode(
      summaryNode,
    (summaryNode.isDir) ?
      DirectoryNodeType.DIRECTORY :
      DirectoryNodeType.FILENAME,
    summaryNode.children.length > 0,
    onExpand,
  );
  summaryNode.children.forEach((childSummaryNode) => {
    root.rows.push(mergeSubTree(childSummaryNode, onExpand));
  });
  return root;
}

function createCoverageMap(
    metrics: CoverageMetric[],
): Map<MetricType, MetricData> {
  const ret = new Map<MetricType, MetricData>();
  metrics.map((metric: CoverageMetric) => {
    if (metric.name === MetricType.LINE.toLowerCase()) {
      let percentageCovered = 100;
      if (metric.total > 0) {
        percentageCovered = metric.covered * 100 / metric.total;
      }

      const metricData: MetricData = {
        covered: metric.covered,
        total: metric.total,
        percentageCovered,
      };
      ret.set('LINE' as MetricType, metricData);
    }
  });

  return ret;
}
