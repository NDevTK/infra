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
  coverageTool: string,
  uiName: string,
  availableRevision: string,
  avaialbleModifierId: string,
}

export interface GetProjectDefaultConfigRequest {
  luci_project: string,
  revision: string,
  modifier_id: string,
}

export interface GetProjectDefaultConfigResponse {
  host: string,
  defaultPlatform: string,
  project: string,
  ref: string,
  platforms: Platform[]
  revision: string,
  modifierId: string,
}

export interface GetSummaryCoverageRequest {
  gitiles_host: string,
  gitiles_project: string,
  gitiles_ref: string,
  gitiles_revision: string,
  path: string,
  unit_tests_only: boolean,
  data_type: string,
  bucket: string,
  builder: string
}

export interface GetSummaryCoverageResponse {
  summary: Summary
}

export interface GetSummaryByComponentRequest {
  gitiles_host: string,
  gitiles_project: string,
  gitiles_ref: string,
  gitiles_revision: string,
  components: string[],
  unit_tests_only: boolean,
  bucket: string,
  builder: string
}

export interface GetSummaryByComponentsResponse {
  summary: Summary[],
}

export interface Team {
  id: string,
  name: string,
  components: string[]
}

export interface GetTeamsResponse {
  teams: Team[]
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
  const tree: SummaryNode[] = fixGetSummaryCoverageResponse(resp);
  return tree;
}

export async function getSummaryCoverageByComponent(
    auth: Auth,
    request: GetSummaryByComponentRequest,
):
  Promise<SummaryNode[]> {
  const resp: GetSummaryByComponentsResponse = await prpcClient.call(
      auth,
      'test_resources.Coverage',
      'GetCoverageSummaryByComponents',
      request,
  );
  const tree: SummaryNode[] = fixGetSummaryCoverageByComponentResponse(resp);
  return tree;
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
  (resp.summary.dirs || []).forEach((dir) => {
    nodes.push({
      name: dir.name,
      path: dir.path,
      children: [] as SummaryNode[],
      summaries: dir.summaries,
      isDir: true,
    } as SummaryNode);
  });
  (resp.summary.files || []).forEach((file) => {
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
    resp: GetSummaryByComponentsResponse,
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
