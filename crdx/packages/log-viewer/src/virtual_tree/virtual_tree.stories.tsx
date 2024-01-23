// Copyright 2024 The Chromium Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { Box } from '@mui/material';
import type { Meta, StoryObj } from '@storybook/react';

import { TreeNodeData } from './types';
import { VirtualTree } from './virtual_tree';

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
  title: 'VirtualTree',
  component: VirtualTree,
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
} satisfies Meta<typeof VirtualTree>;

export default meta;
type Story = StoryObj<typeof VirtualTree>;

const treeData: TreeNodeData[] = [
  {
    id: 1,
    name: 'root1',
    children: [
      {
        id: 3,
        name: 'dir1',
        children: [
          {
            id: 7,
            name: 'leafNode1',
            children: [],
          },
          {
            id: 8,
            name: 'leafNode2',
            children: [],
          },
        ],
      },
      {
        id: 4,
        name: 'dir2',
        children: [
          {
            id: 9,
            name: 'leafNode3',
            children: [],
          },
        ],
      },
    ],
  },
  {
    id: 2,
    name: 'root2',
    children: [
      {
        id: 5,
        name: 'dir3',
        children: [
          {
            id: 10,
            name: 'leafNode4',
            children: [],
          },
          {
            id: 11,
            name: 'leafNode5',
            children: [],
          },
        ],
      },
      {
        id: 6,
        name: 'dir4',
        children: [
          {
            id: 12,
            name: 'leafNode6',
            children: [],
          },
        ],
      },
    ],
  },
];

export const Base: Story = {
  args: {
    root: treeData,
  },
};
