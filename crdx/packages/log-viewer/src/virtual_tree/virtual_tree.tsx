// Copyright 2024 The Chromium Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import React, { useCallback, useEffect, useRef, useState } from 'react';
import { ListRange, Virtuoso, VirtuosoHandle } from 'react-virtuoso';

import { TreeNode } from './tree_node';
import {
  SearchOptions,
  SearchTreeMatch,
  TreeData,
  TreeNodeContainerData,
  TreeNodeData,
} from './types';
import {
  depthFirstSearch,
  generateTreeDataList,
  getSubTreeData,
  isWithinIndexRange,
} from './utils';

const INITIAL_TREE_LEVEL = 0;

const DEFAULT_NODE_INDENTATION = 32;

const SEARCH_PATH_SPLITTER = '/';

/**
 * Props for the Virtual Tree Node Container.
 */
export interface VirtualTreeNodeContainerProps<T extends TreeNodeData> {
  data: TreeNodeContainerData<T>;
  index: number;
  style: React.CSSProperties;
}

export interface VirtualTreeNodeActions<T extends TreeNodeData> {
  onNodeSelect?: (node: TreeData<T>) => void;
  onNodeToggle?: (node: TreeData<T>) => void;
  isSelected?: boolean;
  isSearchMatch?: boolean;
  isActiveSelection?: boolean;
}

/**
 * Props for the Virtual Tree.
 */
export interface VirtualTreeProps<T extends TreeNodeData> {
  /* Data Options */
  root: readonly TreeNodeData[];
  onNodeSelect?: (treeNodeData: T) => void;
  onNodeToggle?: (treeNodeData: T) => void;

  /* Renderers */
  collapseIcon?: React.ReactNode;
  expandIcon?: React.ReactNode;
  // Custom node renderer for the virtual tree.
  itemContent?: (
    index: number,
    row: TreeData<T>,
    context: VirtualTreeNodeActions<T>,
  ) => JSX.Element;
  isTreeCollapsed?: boolean;
  disableVirtualization?: boolean;
  // boolean accessor for activeSelection property of the TreeData which enables
  // rendering of the selected nodes in viewport. If multiple are matched, the
  // first match will rendered in the viewport.
  setActiveSelectionFn?: (treeNodeData: T) => boolean;

  /**
   * Node ids that are marked as selected.
   */
  selectedNodes?: Set<string>;

  /* Search */
  searchOptions?: SearchOptions;

  /**
   * The currently active search index with respect to the total matches. If
   * the prop is not passed, navigation will not be supported.
   */
  searchActiveIndex?: number;

  /**
   * Callback function which returns total search matches when search matches
   * are found.
   */
  onSearchMatchFound?: (totalSearchMatches: number) => void;
}

/**
 * Renders Virtual Tree Component.
 */
export function VirtualTree<T extends TreeNodeData>({
  root,
  searchOptions,
  searchActiveIndex,
  itemContent,
  collapseIcon,
  expandIcon,
  isTreeCollapsed,
  disableVirtualization,
  selectedNodes,
  onSearchMatchFound,
  setActiveSelectionFn,
  onNodeSelect,
  onNodeToggle,
}: VirtualTreeProps<T>) {
  // Map GcsObjectNode id to TreeData<T>.
  const idToTreeDataMap = useRef<Map<string, TreeData<T>>>(new Map());
  // Index of the first visible tree node in the view port.
  const firstVisibleIndex = useRef<number>(-1);
  // Index of the last visible tree node in the view port.
  const lastVisibleIndex = useRef<number>(-1);
  // List of all the tree data nodes.
  const allTreeDataList = useRef<Array<TreeData<T>>>(new Array<TreeData<T>>());
  // Store the initial selection nodeId.
  const shouldRenderInitialSelectionRef = useRef(true);
  // Indicates if the node toggle is in progress the tree browser.
  const isNodeToggleInProgressRef = useRef(false);
  // reference to virtuoso component.
  const virtuosoRef = useRef<VirtuosoHandle>(null);
  // List of node ids from open tree data which match with the search data.
  const searchMatchesRef = useRef<SearchTreeMatch[]>([]);
  const activeSelectionRef = useRef<string | undefined>(undefined);

  // List of expanded(open) tree data which will rendered as tree nodes.
  const [openTreeDataList, setOpenTreeDataList] = useState<Array<TreeData<T>>>(
    new Array<TreeData<T>>(),
  );

  // List of collapsed(closed) tree data which will not be rendered
  //  as tree nodes.
  const closedTreeNodeIdToSubTreeIds = useRef<Map<string, Array<string>>>(
    new Map<string, Array<string>>(),
  );

  // Triggered everytime the list items change continuously updating
  // the first and last visible index in the list.
  const handleOnRangeChanged = ({ startIndex, endIndex }: ListRange) => {
    firstVisibleIndex.current = startIndex;
    lastVisibleIndex.current = endIndex;
  };

  // Callback that scrolls the specific entry at index into view, specialized
  // for variable sized lists
  const scrollEntryIntoViewIfExists = useCallback(
    (searchActiveRowId: number) => {
      if (searchActiveRowId < 0) {
        return;
      }

      const treeNode = openTreeDataList[searchActiveRowId];
      if (treeNode) {
        activeSelectionRef.current = treeNode.id;
      }

      if (
        !isWithinIndexRange(
          searchActiveRowId,
          firstVisibleIndex.current,
          lastVisibleIndex.current,
        )
      ) {
        virtuosoRef.current?.scrollToIndex({
          index: searchActiveRowId,
          align: 'center',
          behavior: 'auto',
        });
      }
    },
    [openTreeDataList],
  );

  /**
   * Updates the open tree data by rendering only expanded tree nodes while
   * collapsed nodes are removed from the DOM.
   */
  const updateOpenTreeDataList = useCallback(() => {
    const allClosedSubTree = Array.from(
      closedTreeNodeIdToSubTreeIds.current.values(),
    ).flat();
    setOpenTreeDataList(
      allTreeDataList.current.filter(
        (treeData: TreeData<T>) => !allClosedSubTree.includes(treeData.id),
      ),
    );
  }, []);

  /**
   * Expands the list of nodes provided.
   */
  const expandNodes = useCallback(
    (treeDataList: Array<TreeData<T>>) => {
      for (const treeData of treeDataList) {
        closedTreeNodeIdToSubTreeIds.current.delete(treeData.id);
        treeData.isOpen = true;
      }
      updateOpenTreeDataList();
    },
    [updateOpenTreeDataList],
  );

  /**
   * Expands all the collapsed nodes in the tree.
   */
  const expandAllNodes = useCallback(() => {
    const nodesToBeExpanded = new Array<TreeData<T>>();
    for (const key of closedTreeNodeIdToSubTreeIds.current.keys()) {
      if (idToTreeDataMap.current.has(key)) {
        nodesToBeExpanded.push(idToTreeDataMap.current.get(key) as TreeData<T>);
      }
    }
    expandNodes(nodesToBeExpanded);
  }, [expandNodes]);

  /**
   * Collapses the list of nodes provided.
   */
  const collapseNodes = useCallback(
    (treeDataList: Array<TreeData<T>>) => {
      for (const treeData of treeDataList) {
        const subTreeDataIdList = getSubTreeData(
          treeData,
          [],
          idToTreeDataMap.current,
        );
        closedTreeNodeIdToSubTreeIds.current.set(
          treeData.id,
          subTreeDataIdList,
        );
        treeData.isOpen = false;
      }
      updateOpenTreeDataList();
    },
    [updateOpenTreeDataList],
  );

  // TODO(b/289439478): fine tune the Fusion View UI for treeBrowser section.
  /**
   * Collapse all nodes in the tree.
   */
  const collapseAllNodes = useCallback(() => {
    const allRootNodes = allTreeDataList.current.filter(
      (treeNode: TreeData<T>) => treeNode.level === 0,
    );
    collapseNodes(allRootNodes);
  }, [collapseNodes]);

  /**
   * Retrieves all the matching search term in the open tree list. Since
   * the virtual tree is not completely rendered, this method returns all
   * the matching search term nodes to scroll into.
   */
  const getSearchMatches = useCallback(() => {
    const searchMatches: SearchTreeMatch[] = [];
    if (!searchOptions) return searchMatches;

    // Dont trigger the search match if the pattern or expanded tree is empty.
    if (!searchOptions.pattern || openTreeDataList.length === 0)
      return searchMatches;

    // Split the search term into non empty regexps so the tree can be searched
    // in the exact order specified by the search path.
    const searchFlag = searchOptions.ignoreCase ? 'i' : '';
    const searchRegexps = getValidRegexps(searchOptions.pattern, searchFlag);
    root.forEach((node) =>
      depthFirstSearch(node, searchRegexps, /* index= */ 0, searchMatches),
    );
    return searchMatches;
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [JSON.stringify(searchOptions), openTreeDataList]);

  // Finds the row Id of the selected search index, this is used to scroll to
  // and highlight the currently active entry.
  const navigateToSearchMatch = useCallback(
    (index: number) => {
      const searchActiveRowId = openTreeDataList.findIndex(
        (treeNode: TreeData<T>) =>
          treeNode.id === searchMatchesRef.current.at(index)?.nodeId,
      );
      scrollEntryIntoViewIfExists(searchActiveRowId);
    },
    [openTreeDataList, scrollEntryIntoViewIfExists],
  );

  /**
   * Toggles the state of the tree node to open or close.
   */
  const handleNodeToggle = (treeData: TreeData<T>) => {
    isNodeToggleInProgressRef.current = true;
    onNodeToggle?.(treeData.data);
    if (treeData.isOpen) {
      collapseNodes([treeData]);
    } else {
      expandNodes([treeData]);
    }
  };

  const handleNodeSelect = (treeData: TreeData<T>) => {
    onNodeSelect?.(treeData.data);
  };

  // Scrolls node into view marked for active selection.
  const renderInitialSelection = useCallback(() => {
    const node = openTreeDataList.find(
      (treeData) => treeData.id === activeSelectionRef.current,
    );
    if (node) {
      const scrollNodeIndex = openTreeDataList.findIndex(
        (treeData) => treeData.id === node.id,
      );
      shouldRenderInitialSelectionRef.current = false;
      virtuosoRef.current?.scrollToIndex({
        index: scrollNodeIndex,
        align: 'center',
        behavior: 'auto',
      });
    }
  }, [openTreeDataList]);

  // Updates the parent component with the search matches and scroll
  // to the first search term if search term is valid.
  const searchTree = useCallback(() => {
    // Reset the node states every time pattern changes to avoid staleness.
    searchMatchesRef.current = [];
    searchMatchesRef.current = getSearchMatches();
    onSearchMatchFound?.(searchMatchesRef.current.length);
  }, [getSearchMatches, onSearchMatchFound]);

  const getValidRegexps = (pattern: string, flag: string): RegExp[] => {
    if (pattern.length === 0) return [];

    // Construct a subtree path if there is a path splitter with valid regexps
    // otherwise default to a single node search.
    let isSubtreeValid = true;
    const regexps: RegExp[] = [];
    for (const segment of pattern.split(SEARCH_PATH_SPLITTER)) {
      if (segment.length === 0) continue;
      try {
        regexps.push(new RegExp(segment, flag));
      } catch {
        isSubtreeValid = false;
        break;
      }
    }

    if (isSubtreeValid) return regexps;

    // Invalid subtree so we attempt to treat the pattern as a single node.
    try {
      return [new RegExp(pattern, flag)];
    } catch {
      return [];
    }
  };

  /**
   * Updates the search state every time the search pattern is updated. The
   * tree state is reset by expanding all the nodes and then the search is
   * performed. However, the tree state can be changed post search. If the
   * search pattern is empty, the search state is reset.
   */
  useEffect(() => {
    if (!searchOptions) return;
    // Expand all the collapsed nodes before searching.
    expandAllNodes();
    if (!searchOptions.pattern) {
      // reset activeSelection and searchMatched fields for all nodes in
      // searchMatched.
      searchMatchesRef.current = [];
    }
    searchTree();
  }, [
    searchOptions?.pattern,
    searchOptions?.ignoreCase,
    searchOptions,
    expandAllNodes,
    searchTree,
  ]);

  /**
   * Recomputes the all tree data list every time the root node
   * changes.
   */
  useEffect(() => {
    if (root.length === 0) return;

    allTreeDataList.current = generateTreeDataList(
      root as T[],
      new Array<TreeData<T>>(),
      INITIAL_TREE_LEVEL,
      undefined,
    );
    allTreeDataList.current.forEach((treeData: TreeData<T>) =>
      idToTreeDataMap.current.set(treeData.id.toString(), treeData),
    );

    // set activeSelection node
    allTreeDataList.current.forEach((treeData) => {
      if (setActiveSelectionFn?.(treeData.data)) {
        activeSelectionRef.current = treeData.id;
      }
    });
    setOpenTreeDataList(allTreeDataList.current);
  }, [root, setActiveSelectionFn]);

  /**
   * Updates the search state of the tree everytime the open tree data or
   * expanded tree list is mutated.
   */
  useEffect(() => {
    // Node toggle updates the open tree data list as the final step and should
    // not trigger rendering of initial selection or search. Using a toggle in
    // progress indicator to capture the same and return.
    if (isNodeToggleInProgressRef.current) {
      isNodeToggleInProgressRef.current = false;
      return;
    }

    // Avoid rendering of initial selection on every re-render.
    if (shouldRenderInitialSelectionRef.current) {
      renderInitialSelection();
    }
    if (!searchOptions?.pattern) return;
    searchTree();
  }, [
    openTreeDataList,
    renderInitialSelection,
    searchOptions?.pattern,
    searchTree,
  ]);

  useEffect(() => {
    if (isTreeCollapsed) {
      collapseAllNodes();
    } else {
      expandAllNodes();
    }
  }, [collapseAllNodes, expandAllNodes, isTreeCollapsed]);

  // Navigate to the next search term and highlight it as active selection.
  useEffect(() => {
    if (searchActiveIndex === undefined || searchActiveIndex < 0) {
      activeSelectionRef.current = undefined;
      return;
    }

    navigateToSearchMatch(searchActiveIndex);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [navigateToSearchMatch, searchActiveIndex, searchMatchesRef.current]);

  // Check if the default node renderer needs to be overridden by user provided
  // custom node renderer.
  const itemContents = (index: number, row: TreeData<T>) => {
    return itemContent ? (
      itemContent(index, row, {
        onNodeSelect: handleNodeSelect,
        onNodeToggle: handleNodeToggle,
        isSelected: selectedNodes?.has(row.id),
        isSearchMatch: searchMatchesRef.current.some(
          (match) => match.nodeId === row.id,
        ),
        isActiveSelection:
          activeSelectionRef.current === row.id && !!searchOptions?.pattern,
      })
    ) : (
      <div
        style={{
          marginLeft: `${DEFAULT_NODE_INDENTATION * row.level}px`,
        }}
      >
        <TreeNode
          data={row}
          onNodeSelect={handleNodeSelect}
          onNodeToggle={handleNodeToggle}
          collapseIcon={collapseIcon}
          expandIcon={expandIcon}
          isSelected={selectedNodes?.has(row.id)}
          isSearchMatch={searchMatchesRef.current.some(
            (match) => match.nodeId === row.id,
          )}
          isActiveSelection={
            activeSelectionRef.current === row.id && !!searchOptions?.pattern
          }
        />
      </div>
    );
  };

  return (
    <Virtuoso
      data-testid={'virtual-tree'}
      ref={virtuosoRef}
      data={openTreeDataList}
      totalCount={openTreeDataList.length}
      rangeChanged={handleOnRangeChanged}
      itemContent={itemContents}
      // Setting this to total count disables virtualization
      initialItemCount={disableVirtualization ? openTreeDataList.length : 0}
      key={disableVirtualization ? openTreeDataList.length : undefined}
    />
  );
}
