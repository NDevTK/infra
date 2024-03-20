// Copyright 2024 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { Stack, Typography } from '@mui/material';
import { ReactNode } from 'react';

import {
  TreeData,
  ObjectNode,
  TreeFontVariant,
  TreeNodeColors,
} from '../types';
import { getNodeBackgroundColor } from '../utils';

/**
 * Props for the Tree node.
 */
interface TreeInternalNodeProps {
  treeNodeData: TreeData<ObjectNode>;
  collapseIcon?: ReactNode;
  expandIcon?: ReactNode;
  treeFontSize?: TreeFontVariant;
  isSearchMatch?: boolean;
  isActiveSelection?: boolean;
  colors?: TreeNodeColors;
  onNodeToggle: (treeNodeData: TreeData<ObjectNode>) => void;
  onNodeSelect: (treeNodeData: TreeData<ObjectNode>) => void;
}

export function TreeInternalNode({
  treeNodeData,
  collapseIcon,
  expandIcon,
  treeFontSize,
  isSearchMatch,
  isActiveSelection,
  colors,
  onNodeSelect,
  onNodeToggle,
}: TreeInternalNodeProps) {
  return (
    <Stack
      spacing={0.5}
      direction={'row'}
      sx={{ display: 'flex', alignItems: 'center' }}
    >
      <span
        role="button"
        tabIndex={0}
        onClick={() => onNodeToggle(treeNodeData)}
        css={{ display: 'flex', alignItems: 'center' }}
      >
        {treeNodeData.isOpen ? collapseIcon : expandIcon}
      </span>
      <Typography
        component="span"
        fontWeight="bold"
        data-testid={`name-${treeNodeData.name}`}
        variant={treeFontSize ?? /* default value */ 'body1'}
      >
        <span
          role="button"
          tabIndex={0}
          onClick={() => onNodeSelect(treeNodeData)}
          css={{
            backgroundColor: getNodeBackgroundColor(
              colors,
              treeNodeData.data.deeplinked,
              isActiveSelection,
              isSearchMatch,
            ),
          }}
        >
          {decodeURIComponent(treeNodeData.name)}
        </span>
      </Typography>
    </Stack>
  );
}
