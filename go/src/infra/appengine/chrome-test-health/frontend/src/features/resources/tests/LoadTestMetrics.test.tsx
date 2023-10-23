// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {
  MetricType,
  Period,
  SortType,
  TestDateMetricData,
  MetricsDateMap,
  DirectoryNode,
  DirectoryNodeType } from '../../../api/resources';
import { computeDates, dataReducer, getLoadedParentIds } from './LoadTestMetrics';
import { Node, Path, Test } from './TestMetricsContext';

function metricsMap(
    metrics: {[date: string]: [MetricType, number][]},
): MetricsDateMap {
  const ret: MetricsDateMap = {};
  for (const date in metrics) {
    if (Object.hasOwn(metrics, date)) {
      ret[date] = {
        data: metrics[date].map(
            (tuple) => ({ metricType: tuple[0], metricValue: tuple[1] }),
        ),
      };
    }
  }
  return ret;
}

describe('computeDates', () => {
  const table = [
    [Period.DAY, false, new Date('2023-07-11T00:00:00'), ['2023-07-11']],
    [Period.DAY, true, new Date('2023-07-11T00:00:00'),
      ['2023-07-07', '2023-07-08', '2023-07-09', '2023-07-10', '2023-07-11']],
    [Period.WEEK, false, new Date('2023-07-11T00:00:00'), ['2023-07-11']],
    [Period.WEEK, true, new Date('2023-07-11T00:00:00'),
      ['2023-06-13', '2023-06-20', '2023-06-27', '2023-07-04', '2023-07-11']],
  ];
  it.each(table)(
      'period:%p timeline:%p date:%p',
      (period, timeline, date, expected) => {
        expect(computeDates({
          page: 0,
          rowsPerPage: 0,
          filter: '',
          date: date as Date,
          period: period as Period,
          sort: SortType.SORT_NAME,
          ascending: true,
          sortIndex: 0,
          timelineMetric: MetricType.AVG_CORES,
          timelineView: timeline as boolean,
          directoryView: false,
        })).toEqual(expected);
      });
});

function pathNode(
    id: string,
    type: DirectoryNodeType,
    loaded: boolean,
    nodes: Node[] = [],
    onExpand?: (node: Node) => void,
): Path {
  return {
    id: id,
    name: id + ((type === DirectoryNodeType.DIRECTORY) ? '/' : ''),
    isExpandable: true,
    metrics: new Map(),
    onExpand: onExpand,
    path: id,
    type: type,
    loaded: loaded,
    rows: nodes,
  };
}

function testNode(
    id: string,
    fileName: string,
    leaf: boolean,
    nodes: Node[] = [],
): Test {
  return {
    id: id,
    name: id,
    isExpandable: !leaf,
    metrics: new Map(),
    rows: nodes,
    fileName: fileName,
  };
}

describe('getLoadedParentIds', () => {
  it('it returns all loaded parent IDs for both directories and files', () => {
    const nodes: Path[] = [
      pathNode('//chrome', DirectoryNodeType.DIRECTORY, true, [
        pathNode('//chrome/app', DirectoryNodeType.DIRECTORY, true, [
          pathNode('//chrome/app/theme', DirectoryNodeType.DIRECTORY, false),
          pathNode('//chrome/app/app.cc', DirectoryNodeType.FILENAME, true, [
            testNode('app', '//chrome/app/app.cc', false),
          ]),
        ]),
        pathNode('//chrome/chrome.cc', DirectoryNodeType.FILENAME, false),
      ]),
      pathNode('//root.cc', DirectoryNodeType.FILENAME, true, [
        testNode('root', '//root.cc', false),
      ]),
    ];
    const [dirs, files] = getLoadedParentIds(nodes);
    expect(dirs).toHaveLength(2);
    expect(dirs).toContain('//chrome');
    expect(dirs).toContain('//chrome/app');
    expect(files).toHaveLength(2);
    expect(files).toContain('//chrome/app/app.cc');
    expect(files).toContain('//root.cc');
  });
});

describe('merge_test action', () => {
  it('populate tests with a single variant correctly', () => {
    const metrics = metricsMap({
      '2012-01-02': [
        [MetricType.NUM_RUNS, 1],
        [MetricType.NUM_FAILURES, 2],
      ],
    });
    const tests: TestDateMetricData[] = [{
      testId: '12',
      testName: 'name',
      fileName: 'file',
      metrics: metrics,
      variants: [
        {
          suite: 'suite',
          builder: 'builder',
          bucket: 'bucket',
          metrics: metricsMap({
            '2012-01-02': [
              [MetricType.NUM_RUNS, 3],
              [MetricType.NUM_FAILURES, 4],
            ],
          }),
        },
      ],
    }];
    const merged = dataReducer([], { type: 'merge_test', tests });
    expect(merged).toHaveLength(1);
    expect(merged[0].id).toEqual(tests[0].testId);
    expect(merged[0].name).toEqual(tests[0].testName);
    expect(merged[0].metrics.size).toEqual(1);
    expect(merged[0].metrics.get('2012-01-02')?.get(MetricType.NUM_RUNS))
        .toEqual(1);
    expect(merged[0].metrics.get('2012-01-02')?.get(MetricType.NUM_FAILURES))
        .toEqual(2);

    expect(merged[0].rows).toHaveLength(1);
    const v = merged[0].rows[0];
    expect(v.name).toEqual(tests[0].variants[0].bucket +
      '/' +tests[0].variants[0].builder);
    expect(v.subname).toEqual(tests[0].variants[0].suite);
    expect(v.metrics.size).toEqual(1);
    expect(v.metrics.get('2012-01-02')?.get(MetricType.NUM_RUNS)).toEqual(3);
    expect(v.metrics.get('2012-01-02')?.get(MetricType.NUM_FAILURES))
        .toEqual(4);
  });

  it('merge tests into existing state correctly', () => {
    const state: Node[] = [pathNode('foo', DirectoryNodeType.FILENAME, false)];
    const tests: TestDateMetricData[] = [{
      testId: '12',
      testName: 'name',
      fileName: 'file',
      metrics: metricsMap({
        '2012-01-02': [
          [MetricType.NUM_RUNS, 1],
        ],
      }),
      variants: [],
    }];
    const merged = dataReducer(state, {
      type: 'merge_test',
      tests: tests,
      parentId: 'foo',
    });
    expect(merged).toHaveLength(1);
    expect(merged[0].id).toEqual('foo');

    expect(merged[0].rows).toHaveLength(1);
    const t = merged[0].rows[0];
    expect(t.id).toEqual(tests[0].testId);
    expect(t.name).toEqual(tests[0].testName);
    expect(t.metrics.size).toEqual(1);
    expect(t.metrics.get('2012-01-02')?.get(MetricType.NUM_RUNS)).toEqual(1);
  });

  it('return empty node for empty tests returned', () => {
    const tests: TestDateMetricData[] = [];
    const merged = dataReducer([], { type: 'merge_test', tests });
    expect(merged).toHaveLength(0);
  });

  it('populate tests with a multiple dates correctly', () => {
    const metrics = metricsMap({
      '2012-01-02': [
        [MetricType.NUM_RUNS, 1],
        [MetricType.NUM_FAILURES, 2],
      ],
      '2012-01-03': [
        [MetricType.NUM_RUNS, 3],
        [MetricType.NUM_FAILURES, 4],
      ],
    });
    const tests: TestDateMetricData[] = [{
      testId: '12',
      testName: 'name',
      fileName: 'file',
      metrics: metrics,
      variants: [],
    }];
    const merged = dataReducer([], { type: 'merge_test', tests });
    expect(merged).toHaveLength(1);
    expect(merged[0].id).toEqual(tests[0].testId);
    expect(merged[0].name).toEqual(tests[0].testName);
    expect(merged[0].metrics.size).toEqual(2);
    expect(merged[0].metrics.get('2012-01-02')?.get(MetricType.NUM_RUNS))
        .toEqual(1);
    expect(merged[0].metrics.get('2012-01-02')?.get(MetricType.NUM_FAILURES))
        .toEqual(2);
    expect(merged[0].metrics.get('2012-01-03')?.get(MetricType.NUM_RUNS))
        .toEqual(3);
    expect(merged[0].metrics.get('2012-01-03')?.get(MetricType.NUM_FAILURES))
        .toEqual(4);
  });

  it('merge test with footer correctly', () => {
    const metrics = metricsMap({
      '2012-01-02': [
        [MetricType.NUM_RUNS, 1],
        [MetricType.NUM_FAILURES, 2],
      ],
    });
    const state: Node[] = [pathNode('/', DirectoryNodeType.DIRECTORY, false)];
    const tests: TestDateMetricData[] = [{
      testId: '12',
      testName: 'name',
      fileName: 'file',
      metrics: metrics,
      variants: [],
    }];
    const merged = dataReducer(
        state,
        { type: 'merge_test', parentId: '/', tests, footer: <div/> },
    );
    expect(merged[0].footer).toBeDefined();
  });

  it('merge test without footer', () => {
    const metrics = metricsMap({
      '2012-01-02': [
        [MetricType.NUM_RUNS, 1],
        [MetricType.NUM_FAILURES, 2],
      ],
    });
    const state: Node[] = [pathNode('/', DirectoryNodeType.DIRECTORY, false)];
    const tests: TestDateMetricData[] = [{
      testId: '12',
      testName: 'name',
      fileName: 'file',
      metrics: metrics,
      variants: [],
    }];
    const merged = dataReducer(
        state,
        { type: 'merge_test', parentId: '/', tests },
    );
    expect(merged[0].footer).toBeUndefined();
  });
});

describe('merge_dir action', () => {
  it('merge a single root node', () => {
    const nodes: DirectoryNode[] = [{
      id: '/',
      type: DirectoryNodeType.DIRECTORY,
      name: 'src',
      metrics: {},
    }];
    const onExpand = () => {/**/};
    const merged = dataReducer([], { type: 'merge_dir', nodes, onExpand });
    expect(merged).toHaveLength(1);
    expect(merged[0].id).toEqual(nodes[0].id);
    expect(merged[0].name).toEqual(nodes[0].name + '/');
    expect(merged[0].rows).toHaveLength(0);
    expect(merged[0].isExpandable).toEqual(true);
    expect(merged[0].onExpand).toBe(onExpand);
    expect((merged[0] as Path).path).toEqual(nodes[0].id);
    expect((merged[0] as Path).loaded).toEqual(false);
  });

  it('merge a single directory node into existing state', () => {
    const state: Node[] = [pathNode('/', DirectoryNodeType.DIRECTORY, false)];
    const nodes: DirectoryNode[] = [{
      id: '/a',
      type: DirectoryNodeType.FILENAME,
      name: 'a',
      metrics: {},
    }];
    const onExpand = () => {/**/};
    const merged = dataReducer(state, {
      type: 'merge_dir',
      parentId: '/',
      nodes: nodes,
      onExpand: onExpand,
    });
    expect(merged).toHaveLength(1);
    expect(merged[0].rows).toHaveLength(1);
    expect((merged[0] as Path).type).toEqual(DirectoryNodeType.DIRECTORY);
    expect((merged[0] as Path).loaded).toEqual(true);

    const m0n0 = merged[0].rows[0];
    expect(m0n0.id).toEqual(nodes[0].id);
    expect(m0n0.name).toEqual(nodes[0].name);
    expect(m0n0.rows).toHaveLength(0);
    expect(m0n0.isExpandable).toEqual(true);
    expect(m0n0.onExpand).toBe(onExpand);
    expect((m0n0 as Path).type).toEqual(DirectoryNodeType.FILENAME);
    expect((m0n0 as Path).loaded).toEqual(false);
  });

  // Tests for a findNode bug where it wouldn't find the node if the first
  // node had children
  it('merge a single directory node into nested second directory', () => {
    const state: Node[] = [
      pathNode('/a', DirectoryNodeType.DIRECTORY, false, [
        pathNode('/a/a', DirectoryNodeType.DIRECTORY, false),
        pathNode('/a/b', DirectoryNodeType.DIRECTORY, false),
      ]),
      pathNode('/b', DirectoryNodeType.DIRECTORY, false),
    ];
    const nodes: DirectoryNode[] = [{
      id: '/b/a',
      type: DirectoryNodeType.FILENAME,
      name: 'ba',
      metrics: {},
    }];
    const onExpand = () => {/**/};
    const merged = dataReducer(state, {
      type: 'merge_dir',
      parentId: '/b',
      nodes: nodes,
      onExpand: onExpand,
    });
    expect(merged).toHaveLength(2);
    expect(merged[0].rows).toHaveLength(2);
    expect(merged[1].rows).toHaveLength(1);
  });
});

describe('rebuild_state action', () => {
  it('rebuild an unexpanded root node', () => {
    const nodes: DirectoryNode[] = [{
      id: '//chrome',
      type: DirectoryNodeType.DIRECTORY,
      name: '//chrome',
      metrics: {},
    }, {
      id: '//infra',
      type: DirectoryNodeType.DIRECTORY,
      name: '//infra',
      metrics: {},
    }];
    const state = dataReducer([], { type: 'rebuild_state', nodes, tests: [] });

    const expected: Node[] = [
      pathNode('//chrome', DirectoryNodeType.DIRECTORY, false),
      pathNode('//infra', DirectoryNodeType.DIRECTORY, false),
    ];
    expect(state).toEqual(expected);
  });

  it('rebuild an expanded root node', () => {
    const nodes: DirectoryNode[] = [{
      id: '//chrome',
      type: DirectoryNodeType.DIRECTORY,
      name: '//chrome',
      metrics: {},
    }, {
      id: '//chrome/app',
      type: DirectoryNodeType.DIRECTORY,
      name: '//chrome/app',
      metrics: {},
    }, {
      id: '//infra',
      type: DirectoryNodeType.DIRECTORY,
      name: '//infra',
      metrics: {},
    }];
    const state = dataReducer([], { type: 'rebuild_state', nodes, tests: [] });

    const expected: Node[] = [
      pathNode('//chrome', DirectoryNodeType.DIRECTORY, true, [
        pathNode('//chrome/app', DirectoryNodeType.DIRECTORY, false),
      ]),
      pathNode('//infra', DirectoryNodeType.DIRECTORY, false),
    ];
    expect(state).toEqual(expected);
  });

  it('rebuild an expanded root node with tests', () => {
    const nodes: DirectoryNode[] = [{
      id: '//chrome',
      type: DirectoryNodeType.DIRECTORY,
      name: '//chrome',
      metrics: {},
    }, {
      id: '//chrome/app.cc',
      type: DirectoryNodeType.FILENAME,
      name: '//chrome/app.cc',
      metrics: {},
    }, {
      id: '//infra',
      type: DirectoryNodeType.DIRECTORY,
      name: '//infra',
      metrics: {},
    }];
    const tests: TestDateMetricData[] = [{
      testId: 'app',
      testName: 'app',
      fileName: '//chrome/app.cc',
      metrics: {},
      variants: [],
    }];
    const state = dataReducer([], { type: 'rebuild_state', nodes, tests });

    const expected: Node[] = [
      pathNode('//chrome', DirectoryNodeType.DIRECTORY, true, [
        pathNode('//chrome/app.cc', DirectoryNodeType.FILENAME, true, [
          testNode('app', '//chrome/app.cc', false),
        ]),
      ]),
      pathNode('//infra', DirectoryNodeType.DIRECTORY, false),
    ];
    expect(state).toEqual(expected);
  });
});
