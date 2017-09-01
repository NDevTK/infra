# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""This module provides API for pipeline."""

import pipeline as pipeline
from pipeline import handlers as pipeline_handlers
from pipeline import status_ui as pipeline_status_ui

from libs.serializeable_object import SerializableObject

_UNDEFINED = object()


class BasePipeline(pipeline.Pipeline):
  """This serves as an """
  input_type = _UNDEFINED
  output_type = _UNDEFINED

  def OnAbort(self, arg):
    """Called when the pipeline is aborted."""
    pass

  def OnFinalized(self, arg):
    """Called when the pipeline is finalized (completed or aborted)"""
    pass

  def RunImpl(self, arg):
    """Called when the pipeline is executing with the given argument."""
    raise NotImplementedError()

  @property
  def pipeline_status_path(self):
    """Returns an absolute path to look up the status of the pipeline."""
    return '/_ah/pipeline/status?root=%s&auto=false' % self.root_pipeline_id

  def __init__(self, *args, **kwargs):
    super(BasePipeline, self).__init__(
        *self._EncodeInputIntoRealPipelineParameters(args, kwargs))

  def _EncodeInputIntoRealPipelineParameters(self, *args, **kwargs):
    # In creation of the pipeline, serializable input should be encoded.
    # In execution of the pipeline, parameters should not be encoded again.
    assert self.input_type is not _UNDEFINED, 'Input type not defined.'
    assert self.output_type is not _UNDEFINED, 'Output type not defined.'
    if (len(args) == 1 and len(kwargs) == 0 and
        issubclass(self.input_type, SerializableObject)):
      assert isinstance(args[0], self.input_type), 'Expected %s, but got %s' % (
          self.input_type.__name__, type(args[0]).__name__)
      kwargs = args[0].ToDict()
      args = []
    return args, kwargs

  def _DecodeRealPipelineParametersIntoInput(self, *args, **kwargs):
    assert len(args) == 0, ('Expected 0 positional arguments, but got %s' %
                            len(args))
    return self.input_type.FromDict(kwargs)

  def run(self, *args, **kwargs):
    arg = self._DecodeRealPipelineParametersIntoInput(self.args, self.kwargs)
    for v in self.RunImpl(arg):
      yield v

  def send_result_email(self):  # pragma: no cover
    """We override this so it doesn't email on completion."""
    pass

  def run_test(self, *args, **kwargs):  # pragma: no cover
    pass

  def finalized_test(self, *args, **kwargs):  # pragma: no cover
    pass
