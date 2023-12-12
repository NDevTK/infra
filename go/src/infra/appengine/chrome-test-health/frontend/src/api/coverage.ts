// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { Auth } from './auth';
import { prpcClient } from './client';

// ---------------- Interfaces ----------------

export interface CoverageMetric {
  name: string,
  covered: number,
  total: number,
}

export interface SummaryNode {
  name: string,
  path: string,
  summaries: CoverageMetric[],
  children: SummaryNode[],
  isDir: boolean,
}

export interface Summary {
  dirs: SummaryNode[],
  files: SummaryNode[],
  path: string,
  summaries: CoverageMetric[]
}

export interface Platform {
  platform: string,
  bucket: string,
  builder: string,
  uiName: string,
  latestRevision: string,
}

export interface CoverageTrend {
  date: string,
  covered: number,
  total: number
}

export interface AbsoluteCoverageTrend {
  date: string,
  linesCovered: number,
  totalLines: number
}

export interface IncrementalCoverageTrend {
  date: string,
  fileChangesCovered: number,
  totalFileChanges: number
}

export interface GetProjectDefaultConfigRequest {
  luci_project: string,
}

export interface GetProjectDefaultConfigResponse {
  gitilesHost: string,
  gitilesProject: string,
  gitilesRef: string,
  builderConfig: Platform[]
}

export interface GetSummaryCoverageRequest {
  gitiles_host: string,
  gitiles_project: string,
  gitiles_ref: string,
  gitiles_revision: string,
  path?: string,
  components?: string[],
  unit_tests_only: boolean,
  bucket: string,
  builder: string
}

export interface GetSummaryCoverageResponse {
  summary: Summary[]
}

export interface Team {
  id: string,
  name: string,
  components: string[]
}

export interface GetTeamsResponse {
  teams: Team[]
}

export interface GetAbsoluteTrendsRequest {
  bucket: string,
  builder: string,
  unit_tests_only: boolean,
  presets?: string[],
  paths: string[],
  components: string[],
}

export interface GetAbsoluteTrendsResponse {
  reports: AbsoluteCoverageTrend[],
}

export interface GetIncrementalTrendsRequest {
  presets?: string[],
  paths: string[],
  components?: string[],
  unit_tests_only: boolean,
}

export interface GetIncrementalTrendsResponse {
  reports: IncrementalCoverageTrend[],
}

// ---------------- RPC Calls ----------------

export async function getProjectDefaultConfig(
    auth: Auth,
    request: GetProjectDefaultConfigRequest,
):
  Promise<GetProjectDefaultConfigResponse> {
  const resp: GetProjectDefaultConfigResponse = await prpcClient.call(
      auth,
      'test_resources.Coverage',
      'GetProjectDefaultConfig',
      request,
  );
  return resp;
}

export async function getSummaryCoverage(
    auth: Auth,
    request: GetSummaryCoverageRequest,
):
  Promise<SummaryNode[]> {
  const resp: GetSummaryCoverageResponse = await prpcClient.call(
      auth,
      'test_resources.Coverage',
      'GetCoverageSummary',
      request,
  );

  return request.components && request.components.length > 0 ?
  fixGetSummaryCoverageByComponentResponse(resp) :
  fixGetSummaryCoverageResponse(resp);
}

export async function getTeams(auth: Auth): Promise<GetTeamsResponse> {
  const resp: GetTeamsResponse = await prpcClient.call(
      auth,
      'test_resources.Coverage',
      'GetTeams',
      {},
  );
  return resp;
}

export async function getAbsoluteCoverageTrends(
    auth: Auth,
    request: GetAbsoluteTrendsRequest,
):
  Promise<GetAbsoluteTrendsResponse> {
  const resp: GetAbsoluteTrendsResponse = await prpcClient.call(
      auth,
      'test_resources.Coverage',
      'GetAbsoluteCoverageDataOneYear',
      request,
  );
  return resp;
}

export async function getIncrementalCoverageTrends(
    auth: Auth,
    request: GetIncrementalTrendsRequest,
):
  Promise<GetIncrementalTrendsResponse> {
  const resp: GetIncrementalTrendsResponse = await prpcClient.call(
      auth,
      'test_resources.Coverage',
      'GetIncrementalCoverageDataOneYear',
      request,
  );
  return resp;
}

// ---------------- Helper Functions ----------------

function getPathParts(path: string, isDir: boolean) {
  let pathParts = path.split('/');
  // For a dir eg: //a/b/ the path will split to ["", "", "a", "b", ""].
  // For a file eg: //a/b.ext the path will split to ["", "", "a", "b.ext"]
  // The following code removes empty string elements
  // at the beginning and the end leaving us with ["a", "b"]
  // and ["a", "b.ext"] respectively for this example.
  pathParts = (isDir) ? pathParts.slice(2, -1) : pathParts.slice(2);
  for (const [i, part] of pathParts.entries()) {
    if (!isDir && i == pathParts.length - 1) break;
    pathParts[i] = `${part}/`;
  }
  return pathParts;
}

function isPathPresent(
    tree: SummaryNode[],
    path: string,
    isDir: boolean,
): boolean {
  const pathParts = getPathParts(path, isDir);
  let treeItr = tree;
  for (const part of pathParts) {
    let hasPart = false;
    for (const node of treeItr) {
      if (node.name === part) {
        hasPart = true;
        treeItr = node.children;
        break;
      }
    }
    if (!hasPart) return false;
  }
  return true;
}

function addPath(
    tree: SummaryNode[],
    path: string,
    isDir: boolean,
    summaries: CoverageMetric[],
) {
  const pathParts = getPathParts(path, isDir);
  let treeItr = tree;
  let pathSoFar = '//';
  for (const [i, part] of pathParts.entries()) {
    pathSoFar = `${pathSoFar}${part}`;
    let hasPart = false;
    let matchingNode: SummaryNode = {
      name: '',
      path: '',
      isDir: false,
      children: [],
      summaries: [],
    };
    for (const node of treeItr) {
      if (node.name === part) {
        hasPart = true;
        matchingNode = node;
        break;
      }
    }
    if (!hasPart) {
      matchingNode = {
        name: part,
        path: pathSoFar,
        isDir: (isDir || i < pathParts.length - 1) ? true : false,
        children: [] as SummaryNode[],
        summaries: [
          { name: 'line', covered: 0, total: 0 },
        ],
      };
      treeItr.push(matchingNode);
    }

    matchingNode.summaries.forEach((summary) => {
      const filteredSummary = summaries.filter(
          (sum) => sum.name === summary.name,
      )[0];
      summary.covered += filteredSummary.covered;
      summary.total += filteredSummary.total;
    });
    treeItr = matchingNode.children;
  }
}

function fixGetSummaryCoverageResponse(
    resp: GetSummaryCoverageResponse,
): SummaryNode[] {
  const nodes: SummaryNode[] = [];
  (resp.summary[0].dirs || []).forEach((dir) => {
    nodes.push({
      name: dir.name,
      path: dir.path,
      children: [] as SummaryNode[],
      summaries: dir.summaries,
      isDir: true,
    } as SummaryNode);
  });
  (resp.summary[0].files || []).forEach((file) => {
    nodes.push({
      name: file.name,
      path: file.path,
      children: [] as SummaryNode[],
      summaries: file.summaries,
      isDir: false,
    } as SummaryNode);
  });
  return nodes;
}

function fixGetSummaryCoverageByComponentResponse(
    resp: GetSummaryCoverageResponse,
): SummaryNode[] {
  const rootNodes: SummaryNode[] = [];
  resp.summary.forEach((s) => {
    s.dirs = s.dirs || [] as SummaryNode[];
    s.files = s.files || [] as SummaryNode[];

    s.dirs.forEach((dir) => {
      if (!isPathPresent(rootNodes, dir.path, true)) {
        addPath(rootNodes, dir.path, true, dir.summaries);
      }
    });
    s.files.forEach((file) => {
      if (!isPathPresent(rootNodes, file.path, false)) {
        addPath(rootNodes, file.path, false, file.summaries);
      }
    });
  });
  return rootNodes;
}
