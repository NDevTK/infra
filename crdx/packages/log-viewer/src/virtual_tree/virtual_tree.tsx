// Copyright 2024 The Chromium Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import React, {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from 'react';
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
  root: readonly T[];

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
   * Toggle the scroll to the active index.
   */
  scrollToggle?: boolean;

  /**
   * Callback function which returns total search matches when search matches
   * are found. activeIndex is -1 when no search matches are found.
   */
  onSearchMatchFound?: (
    activeIndex: number,
    totalSearchMatches: number,
  ) => void;
  onNodeSelect?: (treeNodeData: T) => void;
  onNodeToggle?: (treeNodeData: T) => void;

  // boolean accessor for activeSelection property of the TreeData which enables
  // rendering of the selected nodes in viewport. If multiple are matched, the
  // first match will rendered in the viewport.
  setActiveSelectionFn?: (treeNodeData: T) => boolean;
}

/**
 * Renders Virtual Tree Component.
 */
export function VirtualTree<T extends TreeNodeData>({
  root,
  searchOptions,
  searchActiveIndex,
  scrollToggle,
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
  // Index of the first visible tree node in the view port.
  const firstVisibleIndex = useRef<number>(-1);
  // Index of the last visible tree node in the view port.
  const lastVisibleIndex = useRef<number>(-1);
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

  const { allTreeDataList, idToTreeDataMap } = useMemo(() => {
    const allTreeDataList = generateTreeDataList(
      root as T[],
      new Array<TreeData<T>>(),
      INITIAL_TREE_LEVEL,
      undefined,
    );

    // Map ObjectNode id to TreeData<T>.
    const idToTreeDataMap: Map<string, TreeData<T>> = new Map(
      allTreeDataList.map((treeData) => [treeData.id.toString(), treeData]),
    );

    return { allTreeDataList, idToTreeDataMap };
  }, [root]);

  // Callback that scrolls the specific entry at index into view, specialized
  // for variable sized lists
  const scrollEntryIntoViewIfExists = useCallback(
    (index: number, treeNode: TreeData<T> | undefined) => {
      if (!treeNode) return;
      activeSelectionRef.current = treeNode.id;

      if (
        !isWithinIndexRange(
          index,
          firstVisibleIndex.current,
          lastVisibleIndex.current,
        )
      ) {
        virtuosoRef.current?.scrollToIndex({
          index,
          align: 'center',
          behavior: 'auto',
        });
      }
    },
    [],
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
      allTreeDataList.filter(
        (treeData: TreeData<T>) => !allClosedSubTree.includes(treeData.id),
      ),
    );
  }, [allTreeDataList]);

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
      if (idToTreeDataMap.has(key)) {
        nodesToBeExpanded.push(idToTreeDataMap.get(key) as TreeData<T>);
      }
    }
    expandNodes(nodesToBeExpanded);
  }, [expandNodes, idToTreeDataMap]);

  /**
   * Collapses the list of nodes provided.
   */
  const collapseNodes = useCallback(
    (treeDataList: Array<TreeData<T>>) => {
      for (const treeData of treeDataList) {
        const subTreeDataIdList = getSubTreeData(treeData, [], idToTreeDataMap);
        closedTreeNodeIdToSubTreeIds.current.set(
          treeData.id,
          subTreeDataIdList,
        );
        treeData.isOpen = false;
      }
      updateOpenTreeDataList();
    },
    [idToTreeDataMap, updateOpenTreeDataList],
  );

  // TODO(b/289439478): fine tune the Fusion View UI for treeBrowser section.
  /**
   * Collapse all nodes in the tree.
   */
  const collapseAllNodes = useCallback(() => {
    const allRootNodes = allTreeDataList.filter(
      (treeNode: TreeData<T>) => !treeNode.isLeafNode,
    );
    collapseNodes(allRootNodes);
  }, [allTreeDataList, collapseNodes]);

  /**
   * Retrieves all the matching search term in the open tree list. Since
   * the virtual tree is not completely rendered, this method returns all
   * the matching search term nodes to scroll into.
   */
  const getSearchMatches = useCallback(
    (treeDataList: TreeData<T>[]) => {
      const searchMatches: SearchTreeMatch[] = [];
      if (!searchOptions) return searchMatches;

      // Dont trigger the search match if the pattern or expanded tree is empty.
      if (!searchOptions.pattern || treeDataList.length === 0)
        return searchMatches;

      // Split the search term into non empty regexps so the tree can be searched
      // in the exact order specified by the search path.
      const searchFlag = searchOptions.ignoreCase ? 'i' : '';
      const searchRegexps = getValidRegexps(searchOptions.pattern, searchFlag);
      root.forEach((node) =>
        depthFirstSearch(node, searchRegexps, /* index= */ 0, searchMatches),
      );
      return searchMatches;
    },
    [root, searchOptions],
  );

  // Finds the row Id of the selected search index, this is used to scroll to
  // and highlight the currently active entry.
  const navigateToSearchMatch = useCallback(
    (index: number) => {
      const searchActiveRowId = openTreeDataList.findIndex(
        (treeNode: TreeData<T>) =>
          treeNode.id === searchMatchesRef.current.at(index)?.nodeId,
      );
      const treeNode = openTreeDataList.at(searchActiveRowId);
      scrollEntryIntoViewIfExists(searchActiveRowId, treeNode);
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

  // Figure out the first visible index within the search matches.
  // If none is visible in the current window, reset to 0.
  // If there are no matches it returns -1 as invalid.
  const getFirstVisibleSearchMatchIndex = useCallback(
    (searchMatches: SearchTreeMatch[], treeDataList: TreeData<T>[]): number => {
      if (searchMatches.length === 0) {
        return -1;
      }

      const start = firstVisibleIndex.current;
      const end = lastVisibleIndex.current;
      for (let i = start; i <= end; i++) {
        const treeData = treeDataList.at(i);
        const index = searchMatches.findIndex(
          (match) => match.nodeId === treeData?.id,
        );

        if (index >= 0) return index;
      }

      // Reset the index to 0 since none of the matches are in the current window.
      return 0;
    },
    [],
  );

  // Updates the parent component with the search matches and scroll
  // to the first search term if search term is valid.
  const searchTree = useCallback(
    (treeDataList: TreeData<T>[]) => {
      // Reset the node states every time pattern changes to avoid staleness.
      searchMatchesRef.current = [];
      const matches = getSearchMatches(treeDataList);
      const index = getFirstVisibleSearchMatchIndex(matches, treeDataList);
      searchMatchesRef.current = matches;
      onSearchMatchFound?.(index, matches.length);
    },
    [getSearchMatches, getFirstVisibleSearchMatchIndex, onSearchMatchFound],
  );

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
    searchTree(allTreeDataList);
  }, [searchOptions, expandAllNodes, searchTree, allTreeDataList]);

  /**
   * Computes all tree data list on initial render.
   */
  useEffect(() => {
    // Scroll to the active selection on initial render.
    for (const [index, treeData] of allTreeDataList.entries()) {
      if (setActiveSelectionFn?.(treeData.data)) {
        scrollEntryIntoViewIfExists(index, treeData);
        break;
      }
    }

    setOpenTreeDataList(allTreeDataList);
  }, [allTreeDataList, scrollEntryIntoViewIfExists, setActiveSelectionFn]);

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

    if (!searchOptions?.pattern) return;
    searchTree(openTreeDataList);
  }, [openTreeDataList, searchOptions?.pattern, searchTree]);

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
  }, [
    navigateToSearchMatch,
    searchActiveIndex,
    searchMatchesRef.current,
    scrollToggle,
  ]);

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
