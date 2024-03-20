// Copyright 2024 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { Tooltip, Link } from '@mui/material';
import { grey } from '@mui/material/colors';

import { TreeNodeColors, TreeNodeLabels, ObjectNode } from '../types';

/**
 * Props for the node text.
 */
interface NodeTextProps {
  node: ObjectNode;
  colors?: TreeNodeColors;
}

function NodeText({ node, colors }: NodeTextProps) {
  const nodeInnerText = decodeURIComponent(node.name);

  if (node.size === 0) {
    return (
      <span
        css={{
          color: colors?.unsupportedColor ?? grey[600],
        }}
      >
        {nodeInnerText}
      </span>
    );
  }

  return <Link>{nodeInnerText}</Link>;
}

/**
 * Props for the node text.
 */
interface LeafNodeTextProps {
  node: ObjectNode;
  hasLeafNodeClick: boolean;
  colors?: TreeNodeColors;
  labels: TreeNodeLabels;
}

/**
 * Represents the leaf node text in the logs tree.
 */
export function LeafNodeText({
  node,
  hasLeafNodeClick,
  colors,
  labels,
}: LeafNodeTextProps) {
  if (hasLeafNodeClick) {
    return node.viewingsupported ? (
      <>{decodeURIComponent(node.name)}</>
    ) : (
      <NodeText node={node} colors={colors} />
    );
  } else {
    return (
      <Tooltip title={labels.nonSupportedLeafNodeTooltip} placement="bottom">
        <NodeText node={node} colors={colors} />
      </Tooltip>
    );
  }
}
