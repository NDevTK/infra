// Copyright 2024 The Chromium Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import ChevronRightIcon from '@mui/icons-material/ChevronRight';
import CloudDownloadOutlinedIcon from '@mui/icons-material/CloudDownloadOutlined';
import CloudOutlinedIcon from '@mui/icons-material/CloudOutlined';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import { Box } from '@mui/material';
import type { Meta, StoryObj } from '@storybook/react';

import { ObjectNode, TreeData } from '../types';

import { LogsTreeNode, LogsTreeNodeProps } from './logs_tree_node';

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
  title: 'LogsTreeNode',
  component: LogsTreeNode,
  decorators: [
    (Story) => (
      <Box
        sx={{
          mb: 1,
          height: '600px',
          width: '100%',
          borderTop: '1px solid #e0e0e0',
          borderBottom: '1px solid #e0e0e0',
        }}
      >
        <Story />
      </Box>
    ),
  ],
  // This component will have an automatically generated Autodocs entry: https://storybook.js.org/docs/writing-docs/autodocs
  tags: ['autodocs'],
} satisfies Meta<typeof LogsTreeNode>;

export default meta;
type Story = StoryObj<typeof LogsTreeNode>;

const internalTreeData: TreeData<ObjectNode> = {
  id: '6',
  level: 1,
  name: 'internalNode1',
  isLeafNode: false,
  data: {
    id: 6,
    name: 'internalNode1',
    children: [],
    size: 0,
  },
  children: [],
  parent: undefined,
  isOpen: true,
};

const leafTreeData: TreeData<ObjectNode> = {
  id: '7',
  level: 2,
  name: 'leafNode1',
  isLeafNode: true,
  data: {
    id: 7,
    name: 'leafNode1',
    children: [],
    size: 1000,
  },
  children: [],
  parent: undefined,
  isOpen: true,
};

const commonArgs: Partial<LogsTreeNodeProps<ObjectNode>> = {
  index: 7,
  collapseIcon: <ExpandMoreIcon sx={{ fontSize: '18px' }} />,
  expandIcon: <ChevronRightIcon sx={{ fontSize: '18px' }} />,
  labels: {
    nonSupportedLeafNodeTooltip: 'Tree node viewing is unsupported!',
    specialNodeInfoTooltip: 'Special node is rendered',
  },
};

export const LeafNode: Story = {
  args: {
    treeNodeData: leafTreeData,
    ...commonArgs,
  },
};

export const InternalNodeOpen: Story = {
  args: {
    treeNodeData: internalTreeData,
    ...commonArgs,
  },
};

export const InternalNodeClosed: Story = {
  args: {
    treeNodeData: { ...internalTreeData, isOpen: false },
    ...commonArgs,
  },
};

export const InternalNodeSearchMatch: Story = {
  args: {
    treeNodeData: internalTreeData,
    isSearchMatch: true,
    ...commonArgs,
  },
};

export const InternalNodeActiveSelected: Story = {
  args: {
    treeNodeData: internalTreeData,
    isActiveSelection: true,
    ...commonArgs,
  },
};

export const LeafNodeSelected: Story = {
  args: {
    treeNodeData: leafTreeData,
    isSelected: true,
    ...commonArgs,
  },
};

export const LeafNodeSpecial: Story = {
  args: {
    treeNodeData: leafTreeData,
    isSpecialNode: true,
    ...commonArgs,
  },
};

export const LeafNodeWithIndent: Story = {
  args: {
    treeNodeData: leafTreeData,
    treeIndentBorder: true,
    ...commonArgs,
  },
};

export const LeafNodeWithInlineActions: Story = {
  args: {
    treeNodeData: leafTreeData,
    inlineActions: (
      <span>
        <CloudOutlinedIcon />
        <CloudDownloadOutlinedIcon />
      </span>
    ),
    ...commonArgs,
  },
};
