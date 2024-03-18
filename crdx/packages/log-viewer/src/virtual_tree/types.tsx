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

/**
 * ObjectNode is a node in the logs browser tree sent by the server.
 * It could reference a GCS object, a RBE-CAS artifact or a directory prefix.
 */
export interface ObjectNode {
  id: number;

  /**
   * Immediate filename or dirname.
   */
  name: string;

  /**
   * The url to the resource, it is only set for files, i.e. leaf nodes.
   */
  url?: string;

  /**
   * Length of the object in Bytes.
   */
  size?: number;

  children: ObjectNode[];

  /**
   * Whether the tree should be deeplinked to this node.
   */
  deeplinked?: boolean;

  /**
   * The deeplink path for the node, its the relative path minus the root.
   */
  deeplinkpath?: string;

  /**
   * Whether the node can be viewed in the logs viewer.
   */
  viewingsupported?: boolean;

  /**
   * Indicates if the node is part of a RBE-CAS artifacts tree.
   */
  isRBECAS?: boolean;

  /**
   * Number of log files in the tree. This value is only set in the root node.
   */
  logsCount?: number;

  // UI specific properties:

  /**
   * Whether the node matched the search term.
   */
  searchMatched?: boolean;

  /**
   * Indicates if the node is selected.
   */
  selected?: boolean;
}
