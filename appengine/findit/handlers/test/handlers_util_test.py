# Copyright 2016 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from handlers import handlers_util
from handlers import result_status
from libs import analysis_status
from model.wf_analysis import WfAnalysis
from model.wf_swarming_task import WfSwarmingTask
from model.wf_try_job import WfTryJob
from waterfall.test import wf_testcase


class HandlersUtilResultTest(wf_testcase.WaterfallTestCase):

  def setUp(self):
    super(HandlersUtilResultTest, self).setUp()
    self.master_name = 'm'
    self.builder_name = 'b'
    self.build_number = 121

  def testGetResultAndFailureResultMapForGroup(self):
    first_analysis = WfAnalysis.Create(self.master_name, self.builder_name,
                                       self.build_number)
    first_analysis.result = {'failures': []}
    first_analysis.failure_result_map = {'compile': 'm/b/121'}
    first_analysis.put()

    second_analysis = WfAnalysis.Create('m2', self.builder_name,
                                        self.build_number)
    second_analysis.failure_group_key = [
        self.master_name, self.builder_name, self.build_number
    ]
    second_analysis.put()

    result, failure_result_map = handlers_util._GetResultAndFailureResultMap(
        'm2', self.builder_name, self.build_number, True)

    self.assertEqual(result, first_analysis.result)
    self.assertEqual(failure_result_map, first_analysis.failure_result_map)

  def testGetResultAndFailureResultMapFromOriginalBuild(self):
    first_analysis = WfAnalysis.Create(self.master_name, self.builder_name,
                                       self.build_number)
    first_analysis.result = {'failures': []}
    first_analysis.failure_result_map = {'compile': 'm/b/121'}
    first_analysis.put()

    second_analysis = WfAnalysis.Create('m2', self.builder_name,
                                        self.build_number)
    second_analysis.failure_group_key = [
        self.master_name, self.builder_name, self.build_number
    ]
    second_analysis.put()

    result, failure_result_map = handlers_util._GetResultAndFailureResultMap(
        'm2', self.builder_name, self.build_number)

    self.assertEqual(result, second_analysis.result)
    self.assertEqual(failure_result_map, second_analysis.failure_result_map)

  def testGetSwarmingTaskInfoNoAnalysis(self):
    data = handlers_util.GetSwarmingTaskInfo(
        self.master_name, self.builder_name, self.build_number)
    self.assertEqual({}, data)

  def testGetSwarmingTaskInfoReturnEmptyIfNoFailureMap(self):
    WfAnalysis.Create(self.master_name, self.builder_name,
                      self.build_number).put()

    data = handlers_util.GetSwarmingTaskInfo(
        self.master_name, self.builder_name, self.build_number)

    self.assertEqual({}, data)

  def testGetSwarmingTaskInfoNoSwarmingTasks(self):
    analysis = WfAnalysis.Create(self.master_name, self.builder_name,
                                 self.build_number)
    analysis.failure_result_map = {
        'step1': {
            'test1': '%s/%s/%s' % (self.master_name, self.builder_name, 120),
            'test2': '%s/%s/%s' % (self.master_name, self.builder_name, 120),
            'test3': '%s/%s/%s' % (self.master_name, self.builder_name, 119),
        }
    }
    analysis.put()

    data = handlers_util.GetSwarmingTaskInfo(
        self.master_name, self.builder_name, self.build_number)

    expected_data = {
        'step1': {
            'swarming_tasks': {
                'm/b/119': {
                    'task_info': {
                        'status': result_status.NO_SWARMING_TASK_FOUND
                    },
                    'all_tests': ['test3']
                },
                'm/b/120': {
                    'task_info': {
                        'status': result_status.NO_SWARMING_TASK_FOUND
                    },
                    'all_tests': ['test1', 'test2']
                }
            }
        }
    }

    self.assertEqual(expected_data, data)

  def testGetSwarmingTaskInfoReturnIfNonSwarming(self):
    analysis = WfAnalysis.Create(self.master_name, self.builder_name,
                                 self.build_number)
    analysis.failure_result_map = {
        'step1': '%s/%s/%s' % (self.master_name, self.builder_name, 120)
    }
    analysis.put()

    data = handlers_util.GetSwarmingTaskInfo(
        self.master_name, self.builder_name, self.build_number)

    expected_data = {
        'step1': {
            'swarming_tasks': {
                'm/b/120': {
                    'task_info': {
                        'status': result_status.NON_SWARMING_NO_RERUN
                    }
                }
            }
        }
    }

    self.assertEqual(expected_data, data)

  def testGetSwarmingTaskInfoIfNoSwarmingTask(self):
    analysis = WfAnalysis.Create(self.master_name, self.builder_name,
                                 self.build_number)
    analysis.failure_result_map = {
        'step1': {
            'test1': '%s/%s/%s' % (self.master_name, self.builder_name, 120),
            'test2': '%s/%s/%s' % (self.master_name, self.builder_name, 120),
            'test3': '%s/%s/%s' % (self.master_name, self.builder_name, 119),
        }
    }
    analysis.put()

    data = handlers_util.GetSwarmingTaskInfo(
        self.master_name, self.builder_name, self.build_number)

    expected_data = {
        'step1': {
            'swarming_tasks': {
                'm/b/119': {
                    'task_info': {
                        'status': result_status.NO_SWARMING_TASK_FOUND
                    },
                    'all_tests': ['test3']
                },
                'm/b/120': {
                    'task_info': {
                        'status': result_status.NO_SWARMING_TASK_FOUND
                    },
                    'all_tests': ['test1', 'test2']
                }
            }
        }
    }
    self.assertEqual(expected_data, data)

  def testGetSwarmingTaskInfo(self):
    analysis = WfAnalysis.Create(self.master_name, self.builder_name,
                                 self.build_number)
    analysis.failure_result_map = {
        'step1 on platform': {
            'PRE_test1':
                '%s/%s/%s' % (self.master_name, self.builder_name, 120),
            'PRE_PRE_test2':
                '%s/%s/%s' % (self.master_name, self.builder_name,
                              self.build_number),
            'test3':
                '%s/%s/%s' % (self.master_name, self.builder_name,
                              self.build_number),
            'test4':
                '%s/%s/%s' % (self.master_name, self.builder_name,
                              self.build_number)
        },
        'step2': {
            'test1':
                '%s/%s/%s' % (self.master_name, self.builder_name,
                              self.build_number)
        }
    }
    analysis.put()

    task0 = WfSwarmingTask.Create(self.master_name, self.builder_name, 120,
                                  'step1 on platform')
    task0.task_id = 'task0'
    task0.status = analysis_status.COMPLETED
    task0.parameters = {'tests': ['test1']}
    task0.tests_statuses = {
        'test1': {
            'total_run': 2,
            'SKIPPED': 2
        },
        'PRE_test1': {
            'total_run': 2,
            'FAILURE': 2
        }
    }
    task0.put()

    task1 = WfSwarmingTask.Create(self.master_name, self.builder_name,
                                  self.build_number, 'step1 on platform')
    task1.task_id = 'task1'
    task1.status = analysis_status.COMPLETED
    task1.parameters = {'tests': ['test2', 'test3', 'test4']}
    task1.tests_statuses = {
        'PRE_PRE_test2': {
            'total_run': 2,
            'FAILURE': 2
        },
        'PRE_test2': {
            'total_run': 2,
            'SKIPPED': 2
        },
        'test2': {
            'total_run': 2,
            'SKIPPED': 2
        },
        'test3': {
            'total_run': 4,
            'SUCCESS': 2,
            'FAILURE': 2
        },
        'test4': {
            'total_run': 6,
            'SUCCESS': 6
        }
    }
    task1.put()

    task2 = WfSwarmingTask.Create(self.master_name, self.builder_name,
                                  self.build_number, 'step2')
    task2.put()

    data = handlers_util.GetSwarmingTaskInfo(
        self.master_name, self.builder_name, self.build_number)

    expected_data = {
        'step1 on platform': {
            'swarming_tasks': {
                'm/b/121': {
                    'task_info': {
                        'status':
                            analysis_status.COMPLETED,
                        'task_id':
                            'task1',
                        'task_url': ('https://chromium-swarm.appspot.com/user'
                                     '/task/task1')
                    },
                    'all_tests': ['PRE_PRE_test2', 'test3', 'test4'],
                    'reliable_tests': ['PRE_PRE_test2'],
                    'flaky_tests': ['test3', 'test4'],
                    'ref_name': 'step1'
                },
                'm/b/120': {
                    'task_info': {
                        'status':
                            analysis_status.COMPLETED,
                        'task_id':
                            'task0',
                        'task_url': ('https://chromium-swarm.appspot.com/user/'
                                     'task/task0')
                    },
                    'all_tests': ['PRE_test1'],
                    'reliable_tests': ['PRE_test1'],
                    'flaky_tests': [],
                    'ref_name': 'step1'
                }
            }
        },
        'step2': {
            'swarming_tasks': {
                'm/b/121': {
                    'task_info': {
                        'status': analysis_status.PENDING
                    },
                    'all_tests': ['PRE_test1'],
                    'ref_name': 'step2'
                }
            }
        }
    }
    self.assertEqual(sorted(expected_data), sorted(data))
