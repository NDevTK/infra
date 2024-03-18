// Copyright 2024 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import InfoOutlinedIcon from '@mui/icons-material/InfoOutlined';
import InsertDriveFileOutlinedIcon from '@mui/icons-material/InsertDriveFileOutlined';
import { Stack, Tooltip, Typography, Link } from '@mui/material';
import {
  deepOrange,
  lightBlue,
  teal,
  yellow,
  grey,
} from '@mui/material/colors';
import prettyBytes from 'pretty-bytes';
import { ReactNode, useRef, useState } from 'react';

import { IndentBorder } from './indent_border';
import { TreeData, TreeNodeData, ObjectNode } from './types';

// Total width of 3 inline action icons of 24px each.
const INLINE_ACTIONS_WIDTH = '72px';

export interface TreeNodeLabels {
  nonSupportedLeafNodeTooltip: string;
  specialNodeInfoTooltip: string;
}

/**
 * Defines the color props used in the tree node. Uses default if not provided.
 */
export interface TreeNodeColors {
  activeSelectionBackgroundColor?: string;
  deepLinkBackgroundColor?: string;
  defaultBackgroundColor?: string;
  searchMatchBackgroundColor?: string;
  unsupportedColor?: string;
}

type TreeFontVariant = 'subtitle1' | 'body2' | 'caption';

/**
 * Props for the Tree node.
 */
interface LogsTreeNodeProps<T extends TreeNodeData> {
  treeNodeData: TreeData<T>;
  index: number;
  collapseIcon?: ReactNode;
  expandIcon?: ReactNode;
  treeFontSize?: TreeFontVariant;
  iconFontSize?: string;
  inlineActions?: ReactNode;
  treeIndentBorder?: boolean;
  treeNodeIndentation: number;
  isSelected?: boolean;
  isSearchMatch?: boolean;
  isActiveSelection?: boolean;
  isSpecialNode?: boolean;
  labels: TreeNodeLabels;
  colors?: TreeNodeColors;
  onNodeToggle: (treeNodeData: TreeData<T>) => void;
  onNodeSelect: (treeNodeData: TreeData<T>) => void;
  logActivityTrigger?: (path: string) => void;
  onLeafNodeClick?: (newlySelectedNode: T, mergeFile?: boolean) => void;
  onUnsupportedLeafNodeClick: (node: T) => void;
}

// Gets background to the node based on the treeData attribute.
function getNodeBackgroundColor(
  colors?: TreeNodeColors,
  isDeepLinked?: boolean,
  isActiveSelection?: boolean,
  isSearchMatch?: boolean,
) {
  if (isDeepLinked) return colors?.deepLinkBackgroundColor ?? teal[100];
  if (isActiveSelection)
    return colors?.activeSelectionBackgroundColor ?? deepOrange[300];
  if (isSearchMatch) return colors?.searchMatchBackgroundColor ?? yellow[400];
  return '';
}

function NodeText({
  node,
  colors,
}: {
  node: ObjectNode;
  colors?: TreeNodeColors;
}) {
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

function Node({
  node,
  hasLeafNodeClick,
  colors,
  labels,
}: {
  node: ObjectNode;
  hasLeafNodeClick: boolean;
  colors?: TreeNodeColors;
  labels: TreeNodeLabels;
}) {
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

function TreeLeafNode({
  treeNodeData,
  treeFontSize,
  iconFontSize,
  isSearchMatch,
  isActiveSelection,
  isSpecialNode,
  isHovered,
  inlineActions,
  labels,
  colors,
  onLeafNodeClick,
  onMouseEnter,
  onMouseLeave,
}: {
  treeNodeData: TreeData<ObjectNode>;
  treeFontSize?: TreeFontVariant;
  iconFontSize?: string;
  isSearchMatch?: boolean;
  isActiveSelection?: boolean;
  isSpecialNode?: boolean;
  isHovered?: boolean;
  inlineActions?: ReactNode;
  labels: TreeNodeLabels;
  colors?: TreeNodeColors;
  onLeafNodeClick?: (
    newlySelectedNode: ObjectNode,
    mergeFile?: boolean,
  ) => void;
  onMouseEnter: () => void;
  onMouseLeave: () => void;
}) {
  const nodeDataTestId = onLeafNodeClick
    ? `name-${treeNodeData.name}-with-leaf-handler`
    : `name-${treeNodeData.name}`;
  // Checks if the file is supported.
  const textColorStyle = !treeNodeData.data.viewingsupported
    ? { color: colors?.unsupportedColor ?? grey[600] }
    : {};
  return (
    <Stack direction="row" spacing={1}>
      <Typography
        variant={treeFontSize ?? /* default value */ 'body1'}
        component="span"
        sx={{
          display: 'flex',
          flexWrap: 'nowrap',
          flexDirection: 'row',
          ...textColorStyle,
        }}
      >
        <InsertDriveFileOutlinedIcon
          sx={{ fontSize: iconFontSize ?? '18px' }}
        />
        <span>
          <span
            data-testid={nodeDataTestId}
            css={{
              backgroundColor: getNodeBackgroundColor(
                colors,
                treeNodeData.data.deeplinked,
                isActiveSelection,
                isSearchMatch,
              ),
            }}
          >
            <Node
              node={treeNodeData.data}
              hasLeafNodeClick={!!onLeafNodeClick}
              colors={colors}
              labels={labels}
            />
          </span>
          <span data-testid={`size-${treeNodeData.name}`}>
            {` [${prettyBytes(treeNodeData.data.size!)}]`}
          </span>
        </span>
        {isSpecialNode && (
          <InfoOutlinedIcon
            color="warning"
            sx={{ fontSize: iconFontSize ?? '18px', ml: 0.5 }}
            titleAccess={labels.specialNodeInfoTooltip}
          />
        )}
        {/* Adding this span with fixed width accounts for the width of the
     inline actions. This prevents rearranging of the long format texts
     on hover. */}
        <span
          css={{
            minWidth: INLINE_ACTIONS_WIDTH,
            maxWidth: INLINE_ACTIONS_WIDTH,
            display: 'flex',
            flexDirection: 'row',
            marginLeft: '5px',
          }}
          onMouseEnter={onMouseEnter}
          onMouseLeave={onMouseLeave}
        >
          {isHovered && inlineActions}
        </span>
      </Typography>
    </Stack>
  );
}

function TreeInternalNode({
  treeNodeData,
  collapseIcon,
  expandIcon,
  treeFontSize,
  isSearchMatch,
  isActiveSelection,
  colors,
  onNodeSelect,
  onNodeToggle,
}: {
  treeNodeData: TreeData<ObjectNode>;
  collapseIcon?: ReactNode;
  expandIcon?: ReactNode;
  treeFontSize?: TreeFontVariant;
  isSearchMatch?: boolean;
  isActiveSelection?: boolean;
  colors?: TreeNodeColors;
  onNodeToggle: (treeNodeData: TreeData<ObjectNode>) => void;
  onNodeSelect: (treeNodeData: TreeData<ObjectNode>) => void;
}) {
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

/** Logs tree node representing a file/dir in the directory tree.  */
export function LogsTreeNode({
  treeNodeData,
  index,
  collapseIcon,
  treeFontSize,
  iconFontSize,
  expandIcon,
  inlineActions,
  treeIndentBorder,
  treeNodeIndentation,
  isSelected,
  isSearchMatch,
  isActiveSelection,
  isSpecialNode,
  labels,
  colors,
  onNodeToggle,
  onNodeSelect,
  logActivityTrigger,
  onLeafNodeClick,
  onUnsupportedLeafNodeClick,
}: LogsTreeNodeProps<ObjectNode>) {
  // Reference to the node
  const nodeRef = useRef<HTMLDivElement>(null);

  // TreeData for the node.
  const [isHovered, setIsHovered] = useState(false);
  const [isInlineActionsHovered, setIsInlineActionsHovered] =
    useState<boolean>(false);

  const handleNodeOnClick = (node: ObjectNode) => {
    // Disable the click when the user is hovering over the inline actions.
    if (isInlineActionsHovered) return;

    logActivityTrigger?.(node.deeplinkpath ?? '');
    onNodeSelect(treeNodeData);

    if (onLeafNodeClick && node.viewingsupported) {
      onLeafNodeClick?.(node);
    } else {
      onUnsupportedLeafNodeClick?.(node);
    }
  };

  const backgroundStyles = {
    background: colors?.defaultBackgroundColor ?? lightBlue[50],
    borderRadius: '10px',
  };

  // Adds background color for selected nodes.
  const selectedNodeStyle = isSelected ? backgroundStyles : {};

  // On hover highlights the node.
  const highlightOnHover = isHovered ? backgroundStyles : {};

  return (
    <div
      role="button"
      tabIndex={0}
      ref={nodeRef}
      onClick={() => handleNodeOnClick(treeNodeData.data)}
      data-testid={`node-${treeNodeData.name}`}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
      css={{
        // flex -> required for border indentations.
        // inline-table -> recognizes and adds background color for
        // the extra white space added by browser after rendering text.
        display: `${treeIndentBorder ? 'flex' : 'inline-table'}`,
        flexWrap: 'wrap',
        alignContent: 'center',
        paddingLeft: `${
          treeIndentBorder ? 0 : treeNodeData.level * treeNodeIndentation!
        }px`,
        width: '100%',
        cursor: treeNodeData.data.size === 0 ? 'default' : 'pointer',
        ...selectedNodeStyle,
        ...highlightOnHover,
      }}
    >
      {/* Renders border lines from parent to child */}
      {treeIndentBorder ? (
        <IndentBorder
          index={index}
          level={treeNodeData.level}
          nodeIndentation={treeNodeIndentation}
        />
      ) : (
        <></>
      )}
      {/* Leaf nodes are files. */}
      {treeNodeData.isLeafNode ? (
        <TreeLeafNode
          treeNodeData={treeNodeData}
          treeFontSize={treeFontSize}
          iconFontSize={iconFontSize}
          isSearchMatch={isSearchMatch}
          isActiveSelection={isActiveSelection}
          isHovered={isHovered}
          isSpecialNode={isSpecialNode}
          inlineActions={inlineActions}
          colors={colors}
          labels={labels}
          onLeafNodeClick={onLeafNodeClick}
          onMouseEnter={() => setIsInlineActionsHovered(true)}
          onMouseLeave={() => setIsInlineActionsHovered(false)}
        />
      ) : (
        <TreeInternalNode
          treeNodeData={treeNodeData}
          collapseIcon={collapseIcon}
          expandIcon={expandIcon}
          treeFontSize={treeFontSize}
          isActiveSelection={isActiveSelection}
          isSearchMatch={isSearchMatch}
          colors={colors}
          onNodeSelect={onNodeSelect}
          onNodeToggle={onNodeToggle}
        />
      )}
    </div>
  );
}
