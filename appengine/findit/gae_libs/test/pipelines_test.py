# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import mock

from google.appengine.api import taskqueue

from libs.serializable_object import SerializableObject
from gae_libs import pipelines
from gae_libs.pipelines import BasePipeline
from gae_libs.testcase import TestCase


class _SimpleInfo(SerializableObject):
  param = int


class _ComplexInfo(SerializableObject):
  a = int
  b = _SimpleInfo


class _SyncPipelineWrongOutputType(pipelines.SyncPipeline):
  input_type = int
  output_type = dict

  def RunImpl(self, arg):
    return [arg]


class _SyncPipelineWithBuiltInOutputType(pipelines.SyncPipeline):
  input_type = int
  output_type = list

  def RunImpl(self, arg):
    return [arg]


class _SyncPipelineWithSimpleInfoAsOutputType(pipelines.SyncPipeline):
  input_type = int
  output_type = _SimpleInfo

  def RunImpl(self, arg):
    return _SimpleInfo(param=arg)


class _SyncPipelineWithComplexInfoAsInputType(pipelines.SyncPipeline):
  input_type = _ComplexInfo
  output_type = int

  def RunImpl(self, arg):
    return arg.a + arg.b.param


class _GenPipelineNotSpawnOtherPipelines(pipelines.GenPipeline):
  input_type = int
  output_type = list

  def RunImpl(self, arg):
    return [arg]


class _GenPipelineWithSubPipelines(pipelines.GenPipeline):
  input_type = int
  output_type = _ComplexInfo

  def _ComputeRightAway(self):
    return 10000

  def RunImpl(self, arg):
    a = self._ComputeRightAway()
    b = yield _SyncPipelineWithSimpleInfoAsOutputType(arg)
    complex_info = self.CreateInstance(_ComplexInfo, a=a, b=b)
    yield _SyncPipelineWithComplexInfoAsInputType(complex_info)


class _AsyncPipelineReturnAValueInRunImpl(pipelines.AsyncPipeline):
  input_type = int
  output_type = list
  def RunImpl(self, arg):
    return [arg]


class _AsyncPipelineWithWrongOutputType(pipelines.AsyncPipeline):
  input_type = int
  output_type = list

  def RunImpl(self, arg):
    try:
      task = self.get_callback_task(
          params={'a': arg},
          name=self.pipeline_id)
      task.add()
    except taskqueue.TombstonedTaskError:  # pragma: no branch.
      pass

  def callback(self, **kwargs):
    self.Complete(kwargs)


class _GenPipelineSpawnAsyncPipelineWithWrongOutputType(pipelines.GenPipeline):
  input_type = int

  def RunImpl(self, arg):
    yield _AsyncPipelineWithWrongOutputType(arg)


class _AsyncPipelineOutputAList(pipelines.AsyncPipeline):
  input_type = int
  output_type = list

  def RunImpl(self, arg):
    try:
      task = self.get_callback_task(
          params={'a': arg},
          name=self.pipeline_id)
      task.add()
    except taskqueue.TombstonedTaskError:  # pragma: no branch.
      pass

  def callback(self, **kwargs):
    self.Complete([int(kwargs['a'])])


class PipelinesTest(TestCase):
  app_module = pipelines.pipeline_handlers._APP

  def testAssertionForOnlyOnePositionalInputParameter(self):
    cases = [
        (['a', 'b'], {}),
        (['a'], {'key': 'value'}),
    ]
    for args, kwargs in cases:
      with self.assertRaises(AssertionError):
        pipelines._ConvertInputObjectToPipelineParameters(str, args, kwargs)

  def testNoConvertionForSinglePipelineFutureAsInputParameter(self):
    future = pipelines.pipeline.PipelineFuture(['name'])
    args, kwargs = pipelines._ConvertInputObjectToPipelineParameters(
        int, [future], {})
    self.assertListEqual([future], args)
    self.assertDictEqual({}, kwargs)

  def testInputObjectConvertedToPipelineParameters(self):
    arg = _SimpleInfo(param=1)
    args, kwargs = pipelines._ConvertInputObjectToPipelineParameters(
        _SimpleInfo, [arg], {})
    self.assertListEqual([pipelines._ENCODED_PARAMETER_FLAG], args)
    self.assertDictEqual({'param': 1}, kwargs)

  def testInputObjectConvertedToPipelineParametersOnlyOnce(self):
    args, kwargs = pipelines._ConvertInputObjectToPipelineParameters(
        _SimpleInfo, [pipelines._ENCODED_PARAMETER_FLAG], {'param': 1})
    self.assertListEqual([pipelines._ENCODED_PARAMETER_FLAG], args)
    self.assertDictEqual({'param': 1}, kwargs)

  def testAssertionForInputTypeNotMatch(self):
    with self.assertRaises(AssertionError):
      pipelines._ConvertInputObjectToPipelineParameters(int, ['a'], {})

  def testNoConvertionIfPipelineParameterNotSerializable(self):
    args, kwargs = pipelines._ConvertInputObjectToPipelineParameters(
        int, [1], {})
    self.assertListEqual([1], args)
    self.assertDictEqual({}, kwargs)

  def testAssertionForSinglePipelineParameterIfNotFromSerializableObject(self):
    with self.assertRaises(AssertionError):
        pipelines._ConvertPipelineParametersToInputObject(int, [1, 2], {})

  def testConvertPipelineParametersBackToInputObject(self):
    arg = pipelines._ConvertPipelineParametersToInputObject(
        _SimpleInfo, [pipelines._ENCODED_PARAMETER_FLAG], {'param': 1})
    self.assertTrue(isinstance(arg, _SimpleInfo))
    self.assertEqual(1, arg.param)

  def testAssertionNoKeyValuePipelineParameterIfNotFromSerializableObject(self):
    with self.assertRaises(AssertionError):
        pipelines._ConvertPipelineParametersToInputObject(
            dict, [{'param': 1}], {'key': 'value'})

  def testNoConvertionIfPipelineParameterNotFromSerializableObject(self):
    arg = pipelines._ConvertPipelineParametersToInputObject(
        dict, [{'param': 1}], {})
    self.assertDictEqual({'param': 1}, arg)

  def testInputTypeUndefined(self):
    class InputTypeUndefinedPipeline(BasePipeline):
      output_type = dict
    with self.assertRaises(AssertionError):
      InputTypeUndefinedPipeline('a')

  def testInputTypeNotAType(self):
    class InputTypeUndefinedPipeline(BasePipeline):
      input_type = 123
      output_type = dict
    with self.assertRaises(AssertionError):
      InputTypeUndefinedPipeline('a')

  def testInputTypeNotSupported(self):
    class InputTypeNotSupportedPipeline(BasePipeline):
      input_type = file
      output_type = dict
    with self.assertRaises(AssertionError):
      InputTypeNotSupportedPipeline('file')

  def testOutputTypeUndefined(self):
    class OutputTypeUndefinedPipeline(BasePipeline):
      input_type = dict
    with self.assertRaises(AssertionError):
      OutputTypeUndefinedPipeline('a')

  def testOutputTypeNotAType(self):
    class OutputTypeUndefinedPipeline(BasePipeline):
      input_type = dict
      output_type = 'a'
    with self.assertRaises(AssertionError):
      OutputTypeUndefinedPipeline('a')

  def testOutputTypeNotSupported(self):
    class OutputTypeNotSupportedPipeline(BasePipeline):
      input_type = int
      output_type = basestring
    with self.assertRaises(AssertionError):
      OutputTypeNotSupportedPipeline(1)

  def testWrongOutputTypeForReultOfSyncPipeline(self):
    p = _SyncPipelineWrongOutputType(1)
    p.start()
    self.execute_queued_tasks()
    p = pipelines.pipeline.Pipeline.from_id(p.pipeline_id)
    self.assertTrue(p.was_aborted)

  def testBuiltInTypeAsSyncPipelineOutput(self):
    p = _SyncPipelineWithBuiltInOutputType(1)
    p.start()
    self.execute_queued_tasks()
    p = pipelines.pipeline.Pipeline.from_id(p.pipeline_id)
    self.assertFalse(p.was_aborted)
    self.assertListEqual([1], p.outputs.default.value)

  def testSerializeableObjectAsSyncPipelineOutput(self):
    p = _SyncPipelineWithSimpleInfoAsOutputType(1)
    p.start()
    self.execute_queued_tasks()
    p = pipelines.pipeline.Pipeline.from_id(p.pipeline_id)
    self.assertFalse(p.was_aborted)
    self.assertDictEqual({'param': 1}, p.outputs.default.value)

  def testGenPipelineShouldSpawnOtherPipelines(self):
    p = _GenPipelineNotSpawnOtherPipelines(1)
    p.start()
    self.execute_queued_tasks()
    p = pipelines.pipeline.Pipeline.from_id(p.pipeline_id)
    self.assertTrue(p.was_aborted)

  def testGenPipelineWithSubPipelines(self):
    p = _GenPipelineWithSubPipelines(1)
    p.start()
    self.execute_queued_tasks()
    p = pipelines.pipeline.Pipeline.from_id(p.pipeline_id)
    self.assertFalse(p.was_aborted)
    self.assertEqual(10001, p.outputs.default.value)

  def testAsyncPipelineWithWrongOutputType(self):
    p = _GenPipelineSpawnAsyncPipelineWithWrongOutputType(1)
    p.start()
    self.execute_queued_tasks()
    p = pipelines.pipeline.Pipeline.from_id(p.pipeline_id)
    self.assertTrue(p.was_aborted)

  def testAsyncPipelineOutputAList(self):
    p = _AsyncPipelineOutputAList(1)
    p.start()
    self.execute_queued_tasks()
    p = pipelines.pipeline.Pipeline.from_id(p.pipeline_id)
    self.assertFalse(p.was_aborted)
    self.assertListEqual([1], p.outputs.default.value)

  @mock.patch('logging.warning')
  def testWarningLoggedForAsyncPipelineRunImplReturnAValue(self, warning_func):
    p = _AsyncPipelineReturnAValueInRunImpl(1)
    p.start()
    self.execute_queued_tasks()
    p = pipelines.pipeline.Pipeline.from_id(p.pipeline_id)
    p.complete()
    warning_func.assert_called()
