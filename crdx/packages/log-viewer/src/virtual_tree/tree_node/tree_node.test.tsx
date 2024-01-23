// Copyright 2024 The Chromium Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import '@testing-library/jest-dom';

import { render, fireEvent, screen } from '@testing-library/react';

import { TreeData, TreeNodeData } from '../types';

import { TreeNode } from './tree_node';

const simpleLeafNode: TreeNodeData = {
  id: '1',
  name: 'leaf-node',
  children: [],
};

const simpleNode: TreeNodeData = {
  id: '2',
  name: 'node',
  children: [simpleLeafNode],
};

const LEAF_TREE_NODE_DATA: TreeData<TreeNodeData> = {
  id: '1',
  name: 'leaf-node',
  children: [],
  level: 2,
  data: simpleLeafNode,
  isLeafNode: true,
  isOpen: true,
  parent: undefined,
};

const TREE_NODE_DATA: TreeData<TreeNodeData> = {
  id: '2',
  name: 'node',
  children: [LEAF_TREE_NODE_DATA],
  level: 1,
  data: simpleNode,
  isLeafNode: false,
  isOpen: true,
  parent: undefined,
};

describe('<TreeNode />', () => {
  const mockNodeSelectFn = jest.fn();
  const mockNodeToggleFn = jest.fn();
  it('should render the leaf node', () => {
    render(
      <TreeNode
        data={LEAF_TREE_NODE_DATA}
        onNodeSelect={mockNodeSelectFn}
        onNodeToggle={mockNodeToggleFn}
      />,
    );
    const node = screen.getByTestId('default-leaf-node-1');
    fireEvent.click(node);

    expect(mockNodeSelectFn).toHaveBeenCalled();
    expect(screen.queryByTestId('default-tree-node-1')).toBeInTheDocument();
    expect(screen.queryByTestId('default-leaf-node-1')).toBeInTheDocument();
    expect(screen.queryByTestId('default-node-1')).not.toBeInTheDocument();
  });

  it('should render node', () => {
    render(
      <TreeNode
        data={TREE_NODE_DATA}
        onNodeSelect={mockNodeSelectFn}
        onNodeToggle={mockNodeToggleFn}
      />,
    );
    const node = screen.getByTestId('default-tree-node-2');
    const nodeCollapseIcon = screen.getByTestId('ExpandMoreIcon');
    fireEvent.click(node);
    fireEvent.click(nodeCollapseIcon);

    expect(mockNodeSelectFn).toHaveBeenCalled();
    expect(mockNodeToggleFn).toHaveBeenCalled();
    expect(screen.queryByTestId('default-tree-node-2')).toBeInTheDocument();
    expect(screen.queryByTestId('default-leaf-node-2')).not.toBeInTheDocument();
    expect(screen.queryByTestId('default-node-2')).toBeInTheDocument();
  });
});
