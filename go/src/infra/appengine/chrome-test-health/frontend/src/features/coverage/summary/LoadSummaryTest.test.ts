// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { SummaryNode } from '../../../api/coverage';
import {
  DirectoryNodeType,
  MetricData,
  MetricType,
  Node,
  Path,
  dataReducer,
} from './LoadSummary';

export const mockSummaryNodes: SummaryNode[] = [
  {
    'name': 'dir/',
    'path': '//dir/',
    'isDir': true,
    'children': [
      {
        'name': 'dir1/',
        'path': '//dir/dir1/',
        'isDir': true,
        'children': [],
        'summaries': [
          {
            'name': 'line',
            'covered': 20,
            'total': 100,
          },
        ],
      },
      {
        'name': 'dir2/',
        'path': '//dir/dir2/',
        'isDir': true,
        'children': [],
        'summaries': [
          {
            'name': 'line',
            'covered': 45,
            'total': 100,
          },
        ],
      },
    ],
    'summaries': [
      {
        'name': 'line',
        'covered': 65,
        'total': 200,
      },
    ],
  },
  {
    'name': 'file.ext',
    'path': '//file.ext',
    'isDir': false,
    'children': [],
    'summaries': [
      {
        'name': 'line',
        'covered': 70,
        'total': 100,
      },
    ],
  },
];

function pathNode(
    name: string,
    path: string,
    metrics: Map<MetricType, MetricData>,
    isExpandable: boolean,
    loaded: boolean,
    type: DirectoryNodeType,
    nodes: Node[] = [],
): Path {
  return {
    id: name,
    path,
    name,
    metrics,
    isExpandable,
    onExpand: () => {/**/},
    loaded,
    type,
    rows: nodes,
  };
}

function createMetricMap(
    covered: number,
    total: number,
    percentageCovered: number,
): Map<MetricType, MetricData> {
  const map: Map<MetricType, MetricData> = new Map();
  map.set(MetricType.LINE, { covered, total, percentageCovered } as MetricData);
  return map;
}

describe('merge_dir action', () => {
  it('adds the specified path to the tree', () => {
    const tree = [
      pathNode(
          'dir/', '//dir/', createMetricMap(65, 200, 32.5), true, true,
          DirectoryNodeType.DIRECTORY,
          [
            pathNode(
                'dir1/', '//dir/dir1/', createMetricMap(20, 100, 20), true, false,
                DirectoryNodeType.DIRECTORY, [],
            ),
            pathNode(
                'dir2/', '//dir/dir2/', createMetricMap(45, 100, 45), true, false,
                DirectoryNodeType.DIRECTORY, [],
            ),
          ],
      ),
      pathNode(
          'file.ext', '//file.ext', createMetricMap(70, 100, 70), false, false,
          DirectoryNodeType.FILENAME, [],
      ),
    ];

    const onExpand = () => {/**/};
    const summaryNodes: SummaryNode[] = [{
      'name': 'dir3',
      'path': '//dir/dir1/dir3/',
      'isDir': true,
      'children': [],
      'summaries': [
        {
          'name': 'line',
          'covered': 50,
          'total': 100,
        },
      ],
    }];
    const modifiedTree = dataReducer(
        tree,
        {
          type: 'merge_dir',
          summaryNodes: summaryNodes,
          loaded: true,
          onExpand,
          parentId: 'dir1/',
        },
    );
    expect(modifiedTree[0].rows[0].rows).toHaveLength(1);
    const actual = JSON.stringify(modifiedTree[0].rows[0].rows[0]);
    const expected = JSON.stringify(
        pathNode(
            'dir3', '//dir/dir1/dir3/', createMetricMap(50, 100, 50), true, true,
            DirectoryNodeType.DIRECTORY, [],
        ),
    );
    expect(actual).toEqual(expected);
  });
});

describe('build_tree action', () => {
  it('aggregates the summary nodes correctly', () => {
    const onExpand = () => {/**/};
    const tree = dataReducer([], {
      type: 'build_tree',
      summaryNodes: mockSummaryNodes,
      onExpand,
    });
    const expected = [
      pathNode(
          'dir/', '//dir/', createMetricMap(65, 200, 32.5), true, true,
          DirectoryNodeType.DIRECTORY,
          [
            pathNode(
                'dir1/', '//dir/dir1/', createMetricMap(20, 100, 20), true, false,
                DirectoryNodeType.DIRECTORY, [],
            ),
            pathNode(
                'dir2/', '//dir/dir2/', createMetricMap(45, 100, 45), true, false,
                DirectoryNodeType.DIRECTORY, [],
            ),
          ],
      ),
      pathNode(
          'file.ext', '//file.ext', createMetricMap(70, 100, 70), false, false,
          DirectoryNodeType.FILENAME, [],
      ),
    ];
    expect(JSON.stringify(tree)).toEqual(JSON.stringify(expected));
  });
});

describe('clear_dir action', () => {
  it('empties the tree', () => {
    const tree = [
      pathNode(
          'dir/', '//dir/', createMetricMap(65, 200, 32.5), true, true,
          DirectoryNodeType.DIRECTORY,
          [
            pathNode(
                'dir1/', '//dir/dir1/', createMetricMap(20, 100, 20), true, false,
                DirectoryNodeType.DIRECTORY, [],
            ),
            pathNode(
                'dir2/', '//dir/dir2/', createMetricMap(45, 100, 45), true, false,
                DirectoryNodeType.DIRECTORY, [],
            ),
          ],
      ),
      pathNode(
          'file.ext', '//file.ext', createMetricMap(70, 100, 70), false, false,
          DirectoryNodeType.FILENAME, [],
      ),
    ];
    const clearedTree = dataReducer(tree, { type: 'clear_dir' });
    expect(clearedTree).toHaveLength(0);
  });
});
