# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import apache_beam as beam

from apache_beam.pipeline import PipelineOptions


class EventsPipeline(beam.Pipeline):
  """Pipeline that reads options from the command line."""
  def __init__(self):
    super(EventsPipeline, self).__init__(options=PipelineOptions())


class BQRead(beam.io.iobase.Read):
  """Read transform created from a BigQuerySource with convenient defaults."""

  def __init__(self, query, validate=True, coder=None, use_standard_sql=True,
               flatten_results=False):
    """
    Args:
      query: The query to be run. Should specify table in
        `project.dataset.table` form for standard SQL and
        [project:dataset.table] form is use_standard_sql is False.
      See beam.io.BigQuerySource for explanation of remaining arguments.
    """
    source = beam.io.BigQuerySource(query=query, validate=validate, coder=coder,
                                    flatten_results=flatten_results,
                                    use_standard_sql=use_standard_sql)
    super(BQRead, self).__init__(source)


class BQWrite(beam.io.iobase.Write):
  """Write transform created from a BigQuerySink with convenient defaults."""
  def __init__(self, table, dataset='aggregated'):
    sink = beam.io.BigQuerySink(table, dataset, project='chrome-infra-events')
    super(BQWrite, self).__init__(sink)
