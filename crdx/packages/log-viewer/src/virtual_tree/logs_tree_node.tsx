// Copyright 2024 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import InfoOutlinedIcon from '@mui/icons-material/InfoOutlined';
import InsertDriveFileOutlinedIcon from '@mui/icons-material/InsertDriveFileOutlined';
import { Stack, Tooltip, Typography } from '@mui/material';
import { deepOrange, lightBlue, teal, yellow } from '@mui/material/colors';
import prettyBytes from 'pretty-bytes';
import { ReactNode, memo, useEffect, useRef, useState } from 'react';

import { IndentBorder } from './indent_border';
import { TreeData, TreeNodeData, ObjectNode } from './types';

// Total width of 3 inline action icons of 24px each.
const INLINE_ACTIONS_WIDTH = '72px';

export interface TreeNodeLabels {
  nonSupportedLeafNodeTooltip: string;
  specialNodeInfoTooltip: string;
}

/**
 * Props for the Tree node.
 */
interface LogsTreeNodeProps<T extends TreeNodeData> {
  data: TreeData<T>;
  index: number;
  collapseIcon?: ReactNode;
  expandIcon?: ReactNode;
  treeFontSize?: 'subtitle1' | 'body2' | 'caption';
  inlineActions?: JSX.Element;
  treeIndentBorder?: boolean;
  treeNodeIndentation: number;
  isSelected?: boolean;
  isSearchMatch?: boolean;
  isActiveSelection?: boolean;
  isSpecialNode?: boolean;
  labels: TreeNodeLabels;
  onNodeToggle: (treeNodeData: TreeData<T>) => void;
  onNodeSelect: (treeNodeData: TreeData<T>) => void;
  logActivityTrigger?: (path: string) => void;
  onLeafNodeClick?: (newlySelectedNode: T, mergeFile?: boolean) => void;
  onUnsupportedLeafNodeClick: (node: T) => void;
}

/** Logs tree node representing a file/dir in the directory tree.  */
export const LogsTreeNode = memo(function LogsTreeNode({
  data,
  index,
  collapseIcon,
  treeFontSize,
  expandIcon,
  inlineActions,
  treeIndentBorder,
  treeNodeIndentation,
  isSelected,
  isSearchMatch,
  isActiveSelection,
  isSpecialNode,
  labels,
  onNodeToggle,
  onNodeSelect,
  logActivityTrigger,
  onLeafNodeClick,
  onUnsupportedLeafNodeClick,
}: LogsTreeNodeProps<ObjectNode>) {
  // Reference to the node
  const nodeRef = useRef<HTMLDivElement>(null);

  // TreeData for the node.
  const [treeNodeData, setTreeNodeData] = useState<TreeData<ObjectNode>>(data);
  const [isHovered, setIsHovered] = useState(false);
  const [isInlineActionsHovered, setIsInlineActionsHovered] =
    useState<boolean>(false);

  // Gets background to the node based on the treeData attribute.
  const getNodeBackgroundColor = (node: TreeData<ObjectNode>): string => {
    if (node.data.deeplinked) return teal[100];
    if (isActiveSelection) return deepOrange[300];
    if (isSearchMatch) return yellow[400];
    return '';
  };

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

  const getNodeText = (node: ObjectNode): JSX.Element => {
    return (
      <span
        style={
          node.size === 0
            ? { color: 'gray' }
            : { textDecoration: 'underline', color: '-webkit-link' }
        }
      >
        {decodeURIComponent(node.name)}
      </span>
    );
  };

  const renderNodeText = (node: ObjectNode) => {
    if (onLeafNodeClick) {
      return node.viewingsupported ? (
        <>{decodeURIComponent(node.name)}</>
      ) : (
        getNodeText(node)
      );
    } else {
      return (
        <Tooltip title={labels.nonSupportedLeafNodeTooltip} placement="bottom">
          {getNodeText(node)}
        </Tooltip>
      );
    }
  };

  // Update the treeNode data if the data or index property change.
  useEffect(() => {
    setTreeNodeData(data);
  }, [index, data]);

  // Checks if the file is supported.
  const textColorStyle = !treeNodeData.data.viewingsupported
    ? { color: 'gray' }
    : {};
  const nodeDataTestId = onLeafNodeClick
    ? `name-${treeNodeData.name}-with-leaf-handler`
    : `name-${treeNodeData.name}`;
  const backgroundStyles = {
    background: lightBlue[50],
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
      style={{
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
        <Stack direction="row" spacing={1}>
          <Typography
            variant={treeFontSize ?? /* default value */ 'body1'}
            component="span"
            style={{
              display: 'flex',
              flexWrap: 'nowrap',
              flexDirection: 'row',
              ...textColorStyle,
            }}
          >
            <InsertDriveFileOutlinedIcon sx={{ fontSize: '18px' }} />
            <span>
              <span
                data-testid={nodeDataTestId}
                style={{
                  backgroundColor: getNodeBackgroundColor(treeNodeData),
                }}
              >
                {renderNodeText(treeNodeData.data)}
              </span>
              <span data-testid={`size-${treeNodeData.name}`}>
                {` [${prettyBytes(treeNodeData.data.size!)}]`}
              </span>
            </span>
            {isSpecialNode && (
              <InfoOutlinedIcon
                color="warning"
                sx={{ fontSize: '18px', ml: 0.5 }}
                titleAccess={labels.specialNodeInfoTooltip}
              />
            )}
            {/* Adding this span with fixed width accounts for the width of the
             inline actions. This prevents rearranging of the long format texts
             on hover. */}
            <span
              style={{
                minWidth: INLINE_ACTIONS_WIDTH,
                maxWidth: INLINE_ACTIONS_WIDTH,
                display: 'flex',
                flexDirection: 'row',
                marginLeft: '5px',
              }}
              onMouseEnter={() => setIsInlineActionsHovered(true)}
              onMouseLeave={() => setIsInlineActionsHovered(false)}
            >
              {isHovered && inlineActions}
            </span>
          </Typography>
        </Stack>
      ) : (
        <Stack
          spacing={0.5}
          direction={'row'}
          style={{ display: 'flex', alignItems: 'center' }}
        >
          <span
            role="button"
            tabIndex={0}
            onClick={() => onNodeToggle(treeNodeData)}
            style={{ display: 'flex', alignItems: 'center' }}
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
              style={{
                backgroundColor: getNodeBackgroundColor(treeNodeData),
              }}
            >
              {decodeURIComponent(treeNodeData.name)}
            </span>
          </Typography>
        </Stack>
      )}
    </div>
  );
});
