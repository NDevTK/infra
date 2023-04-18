# Copyright 2022 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

try:
  import html
except ImportError:
  import cgi as html
import json
import logging
import time

from infra_libs.ts_mon.common import distribution
from infra_libs.ts_mon.common import interface
from infra_libs.ts_mon.common import metrics


class TSMonJSFlaskHandler(object):
  """Proxy handler for ts_mon metrics collected in JavaScript.

  To use this class:
  1. Subclass it and override self.xsrf_is_valid
  2. After instantiation call self.register_metrics to register global metrics.
  """

  def __init__(self, flask, services=None):
    self.flask = flask
    self._metrics = None
    self.response = None
    self.request = self.flask.request
    self.service = services

  def register_metrics(self, metrics_list):
    """Registers ts_mon metrics, required for use.

    Args:
      metrics_list: a list of definitions, from ts_mon.metrics.
    """
    interface.register_global_metrics(metrics_list)
    self._metrics = self._metrics or {}
    for metric in metrics_list:
      if metric.is_cumulative():
        metric.dangerously_enable_cumulative_set()
      self._metrics[metric.name] = metric

  def post(self):
    """POST expects a JSON body that's a dict which includes a key "metrics".
    This key's value is an array of objects with schema:
    {
      "metrics": [{
        "MetricInfo": {
          "Name": "monorail/frontend/float_test",
          "ValueType": 2
        },
        "Cells": [{
          "value": 1,
          "fields": {},
          "start_time": 1538430628174
        }]
      }]
    }

    Important!
    The user of this library is responsible for validating XSRF tokens via
    implementing the method self.xsrf_is_valid.
    """
    self.response = self.flask.make_response()
    self.response.headers.add('Content-Security-Policy', "default-src 'none'")
    self.response.content_type = 'text/plain; charset=UTF-8'

    if not self._metrics:
      self.response.status_code = 400
      self.response.data = 'No metrics have been registered.'
      logging.warning('gae_ts_mon error: No metrics have been registered.')
      return self.response

    try:
      # TODO: (crbug.com/monorail/10992) update the JSON parse functions
      body = json.loads(self.request.get_data(as_text=True))
    except ValueError:
      self.response.status_code = 400
      self.response.data = 'Invalid JSON.'
      logging.warning('gae_ts_mon error: Invalid JSON.')
      return self.response

    if not self.xsrf_is_valid(body):
      self.response.status_code = 403
      self.response.data = 'XSRF token invalid.'
      logging.warning('gae_ts_mon error: XSRF token invalid.')
      return self.response

    if not isinstance(body, dict):
      self.response.status_code = 400
      self.response.data = 'Body must be a dictionary.'
      logging.warning('gae_ts_mon error: Body must be a dictionary.')
      return self.response

    if 'metrics' not in body:
      self.response.status_code = 400
      self.response.data = 'Key "metrics" must be in body.'
      logging.warning('gae_ts_mon error: Key "metrics" must be in body.')
      logging.warning('Request body: %s', body)
      return self.response

    for metric_measurement in body.get('metrics', []):
      name = metric_measurement.get('MetricInfo').get('Name')
      metric = self._metrics.get(name, None)

      if not metric:
        self.response.status_code = 400
        self.response.data = ('Metric "%s" is not defined.' % html.escape(name))
        logging.warning('gae_ts_mon error: Metric "%s" is not defined.', name)
        return self.response

      for cell in metric_measurement.get('Cells', []):
        fields = cell.get('fields', {})
        value = cell.get('value')

        metric_field_keys = set(fs.name for fs in metric.field_spec)
        if set(fields.keys()) != metric_field_keys:
          self.response.status_code = 400
          self.response.data = ('Supplied fields do not match metric "%s".' %
                                html.escape(name))
          logging.warning(
              'gae_ts_mon error: Supplied fields do not match metric "%s".',
              name)
          logging.warning('Supplied fields keys: %s', list(fields.keys()))
          logging.warning('Metric fields keys: %s', metric_field_keys)
          return self.response

        start_time = cell.get('start_time')
        if metric.is_cumulative() and not start_time:
          self.response.status_code = 400
          self.response.data = 'Cumulative metrics must have start_time.'
          logging.warning(
              'gae_ts_mon error: Cumulative metrics must have start_time.')
          logging.warning('Metric name: %s', name)
          return self.response

        if metric.is_cumulative() and not self._start_time_is_valid(start_time):
          self.response.status_code = 400
          self.response.data = ('Invalid start_time: %s.' %
                                html.escape(str(start_time)))
          logging.warning('gae_ts_mon error: Invalid start_time: %s.',
                          start_time)
          return self.response

        # Convert distribution metric values into Distribution classes.
        if (isinstance(metric, (metrics.CumulativeDistributionMetric,
                                metrics.NonCumulativeDistributionMetric))):
          if not isinstance(value, dict):
            self.response.status_code = 400
            self.response.data = (
                'Distribution metric values must be a dict.')
            logging.warning(
                'gae_ts_mon error: Distribution metric values must be a dict.')
            logging.warning('Metric value: %s', value)
            return self.response
          dist_value = distribution.Distribution(bucketer=metric.bucketer)
          dist_value.sum = value.get('sum', 0)
          dist_value.count = value.get('count', 0)
          dist_value.buckets.update(value.get('buckets', {}))
          metric.set(dist_value, fields=fields)
        else:
          metric.set(value, fields=fields)

        if metric.is_cumulative():
          metric.dangerously_set_start_time(start_time)

    self.response.status_code = 201
    self.response.data = 'Ok.'
    return self.response

  def xsrf_is_valid(self, _body):
    """Takes a request body and returns whether the included XSRF token
    is valid.

    This method must be implemented by a subclass.
    """
    raise NotImplementedError(
        'xsrf_is_valid must be implemented in a subclass.')

  def time_fn(self):
    """Defaults to time.time. Can be overridden for testing."""
    return time.time()

  def _start_time_is_valid(self, start_time):
    """Validates that a start_time is not in the future and not
    more than a month in the past.
    """
    now = self.time_fn()
    if start_time > now:
      return False

    one_month_seconds = 60 * 60 * 24 * 30
    one_month_ago = now - one_month_seconds
    if start_time < one_month_ago:
      return False

    return True
