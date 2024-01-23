// Copyright 2024 The Chromium Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

/**
 * Represents the fields on the tree data source.
 */
export interface TreeNodeData {
  id: string | number;
  name: string;
  children: TreeNodeData[];
}

/**
 * Provides comprehensive search options.
 */
export interface SearchOptions {
  pattern: string;
  enableRegex?: boolean;
  ignoreCase?: boolean;
  filterOnSearch?: boolean;
}

/**
 * Represents tree match data.
 */
export interface SearchTreeMatch {
  nodeId: string;
}

/**
 * Represents the Virtual Node data with additional tree
 * properties.
 */
export interface TreeData<T extends TreeNodeData> {
  id: string;
  level: number;
  name: string;
  isLeafNode: boolean;
  data: T;
  children: Array<TreeData<T>>;
  isOpen: boolean;
  parent: TreeData<T> | undefined;
}

/**
 * Structure for tree node container data.
 */
export interface TreeNodeContainerData<T extends TreeNodeData> {
  handleNodeToggle: (node: TreeData<T>) => void;
  handleNodeSelect: (node: TreeData<T>) => void;
  treeDataList: Array<TreeData<T>>;
  collapseIcon?: React.ReactNode;
  expandIcon?: React.ReactNode;
}
