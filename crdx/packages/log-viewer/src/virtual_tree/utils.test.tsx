// Copyright 2024 The Chromium Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { TreeData, TreeNodeData } from './types';
import {
  generateTreeDataList,
  getSubTreeData,
  isWithinIndexRange,
} from './utils';

const root: TreeNodeData = {
  id: 1,
  name: 'root',
  children: [],
};

const rootTreeData: TreeData<TreeNodeData> = {
  name: 'root',
  id: '1',
  data: { ...root, id: '1' },
  level: 0,
  children: [],
  isLeafNode: false,
  isOpen: true,
  parent: undefined,
};

const leaf1: TreeNodeData = {
  id: 3,
  name: 'leaf1',
  children: [],
};

const leaf1TreeData: TreeData<TreeNodeData> = {
  name: 'leaf1',
  id: '3',
  data: { ...leaf1, id: '3' },
  children: [],
  level: 1,
  isLeafNode: true,
  isOpen: true,
  parent: rootTreeData,
};

const leaf2: TreeNodeData = {
  id: 4,
  name: 'leaf2',
  children: [],
};

const leaf2TreeData: TreeData<TreeNodeData> = {
  name: 'leaf2',
  id: '4',
  data: { ...leaf2, id: '4' },
  level: 1,
  children: [],
  isLeafNode: true,
  isOpen: true,
  parent: rootTreeData,
};

root.children.push(leaf1);
root.children.push(leaf2);
rootTreeData.children.push(leaf1TreeData);
rootTreeData.children.push(leaf2TreeData);

const treeData: TreeNodeData[] = [root];

describe('Virtual Tree utils', () => {
  it('should generate tree data list without active selection accessor', () => {
    const treeDataList = generateTreeDataList(treeData, [], 0, undefined);
    expect(treeDataList).toHaveLength(3);
    expect(treeDataList.map((treeData) => treeData.id)).toEqual([
      '1',
      '3',
      '4',
    ]);
  });

  it('should check if index is within range', () => {
    expect(isWithinIndexRange(2, 0, 3)).toBeTruthy();
    expect(isWithinIndexRange(5, 0, 3)).toBeFalsy();
  });

  it('should return list of subtree ids', () => {
    const subTreeListIds = getSubTreeData(
      rootTreeData,
      [],
      new Map<string, TreeData<TreeNodeData>>([
        ['1', rootTreeData],
        ['3', leaf1TreeData],
        ['4', leaf2TreeData],
      ]),
    );
    expect(subTreeListIds).toEqual(['3', '4']);
  });
});
