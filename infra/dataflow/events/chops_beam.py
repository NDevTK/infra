import apache_beam as beam

class EventsPipeline(beam.Pipeline):
  def __init__(self, additional_argv=None):
    argv = ['--project', 'chrome-infra-events']
    if additional_argv:
      argv += additional_argv
    super(EventsPipeline, self).__init__(argv=argv)


class BQRead(beam.io.iobase.Read):
  """ Initializes a BigQuerySource with convenient defaults and creates a
      standard Read transform from that source.
  """

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
  """ Initializes a BigQuerySink with convenient defaults and creates a
      standard Write transform from that sink.
  """
  def __init__(self, table, dataset='aggregated'):
    sink = beam.io.BigQuerySink(table, dataset, project='chrome-infra-events')
    super(BQWrite, self).__init__(sink)
