# Copyright 2015 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import sys
import unittest

import mock
from infra_libs.ts_mon.protos.current import metrics_pb2
from infra_libs.ts_mon.protos.new import metrics_pb2 as new_metrics_pb2

from infra_libs.ts_mon.common import distribution
from infra_libs.ts_mon.common import errors
from infra_libs.ts_mon.common import interface
from infra_libs.ts_mon.common import metric_store
from infra_libs.ts_mon.common import metrics
from infra_libs.ts_mon.common import targets
from infra_libs.ts_mon.common.test import stubs


class TestBase(unittest.TestCase):
  def setUp(self):
    super(TestBase, self).setUp()
    target = targets.TaskTarget('test_service', 'test_job',
                                'test_region', 'test_host')
    self.mock_state = interface.State(target=target)
    self.state_patcher = mock.patch('infra_libs.ts_mon.common.interface.state',
                                    new=self.mock_state)
    self.state_patcher.start()

  def tearDown(self):
    self.state_patcher.stop()
    super(TestBase, self).tearDown()


class MetricTest(TestBase):

  def test_name_property(self):
    m1 = metrics.Metric('/foo', fields={'asdf': 1})
    self.assertEquals(m1.name, 'foo')

  def test_init_too_may_fields(self):
    fields = {str(i): str(i) for i in xrange(8)}
    with self.assertRaises(errors.MonitoringTooManyFieldsError) as e:
      metrics.Metric('test', fields=fields)
    self.assertEquals(e.exception.metric, 'test')
    self.assertEquals(len(e.exception.fields), 8)

  def test_serialize(self):
    t = targets.DeviceTarget('reg', 'role', 'net', 'host')
    m = metrics.StringMetric('test')
    m.set('val')
    p = metrics_pb2.MetricsCollection()
    m.serialize_to(p, 1234, (('bar', 1), ('baz', False)), m.get(), t)
    return str(p).splitlines()

  def test_serialize_with_description(self):
    t = targets.DeviceTarget('reg', 'role', 'net', 'host')
    m = metrics.StringMetric('test', description='a custom description')
    m.set('val')
    p = metrics_pb2.MetricsCollection()
    m.serialize_to(p, 1234, (('bar', 1), ('baz', False)), m.get(), t)
    return str(p).splitlines()

  def test_serialize_with_units(self):
    t = targets.DeviceTarget('reg', 'role', 'net', 'host')
    m = metrics.GaugeMetric('test', units=metrics.MetricsDataUnits.SECONDS)
    m.set(1)
    p = metrics_pb2.MetricsCollection()
    m.serialize_to(p, 1234, (('bar', 1), ('baz', False)), m.get(), t)
    self.assertEquals(p.data[0].units, metrics.MetricsDataUnits.SECONDS)
    return str(p).splitlines()

  def test_serialize_too_many_fields(self):
    m = metrics.StringMetric('test', fields={'a': 1, 'b': 2, 'c': 3, 'd': 4})
    m.set('val', fields={'e': 5, 'f': 6, 'g': 7})
    with self.assertRaises(errors.MonitoringTooManyFieldsError):
      m.set('val', fields={'e': 5, 'f': 6, 'g': 7, 'h': 8})

  def test_populate_field_values(self):
    pb1 = metrics_pb2.MetricsData()
    m1 = metrics.Metric('foo', fields={'asdf': 1})
    m1._populate_fields(pb1, m1._normalized_fields)
    self.assertEquals(pb1.fields[0].name, 'asdf')
    self.assertEquals(pb1.fields[0].int_value, 1)

    pb2 = metrics_pb2.MetricsData()
    m2 = metrics.Metric('bar', fields={'qwer': True})
    m2._populate_fields(pb2, m2._normalized_fields)
    self.assertEquals(pb2.fields[0].name, 'qwer')
    self.assertEquals(pb2.fields[0].bool_value, True)

    pb3 = metrics_pb2.MetricsData()
    m3 = metrics.Metric('baz', fields={'zxcv': 'baz'})
    m3._populate_fields(pb3, m3._normalized_fields)
    self.assertEquals(pb3.fields[0].name, 'zxcv')
    self.assertEquals(pb3.fields[0].string_value, 'baz')

  def test_invalid_field_value(self):
    pb = metrics_pb2.MetricsData()
    m = metrics.Metric('test', fields={'pi': 3.14})
    with self.assertRaises(errors.MonitoringInvalidFieldTypeError) as e:
      m._populate_fields(pb, m._normalized_fields)
    self.assertEquals(e.exception.metric, 'test')
    self.assertEquals(e.exception.field, 'pi')
    self.assertEquals(e.exception.value, 3.14)

  def test_register_unregister(self):
    self.assertEquals(0, len(self.mock_state.metrics))
    m = metrics.Metric('test', fields={'pi': 3.14})
    self.assertEquals(1, len(self.mock_state.metrics))
    m.unregister()
    self.assertEquals(0, len(self.mock_state.metrics))

  def test_reset(self):
    m = metrics.StringMetric('test')
    self.assertIsNone(m.get())
    m.set('foo')
    self.assertEqual('foo', m.get())
    m.reset()
    self.assertIsNone(m.get())

  def test_map_units(self):
    units = metrics.MetricsDataUnits
    self.assertEqual(
        '{unknown}', metrics.Metric._map_units_to_string(units.UNKNOWN_UNITS))
    self.assertEqual('{unknown}',
                     metrics.Metric._map_units_to_string('random_value'))
    self.assertEqual('s', metrics.Metric._map_units_to_string(units.SECONDS))

  @mock.patch('infra_libs.ts_mon.common.metrics.Metric.is_cumulative',
              autospec=True)
  @mock.patch('infra_libs.ts_mon.common.metrics.Metric._populate_value_type',
              autospec=True)
  @mock.patch(('infra_libs.ts_mon.common.metrics.Metric'
               '._populate_field_descriptors'), autospec=True)
  def test_populate_data_set(self, _value_type, _fd,  _is_cumulative):
    interface.state.metric_name_prefix = '/infra/test/'
    scenarios = [(True, 'desc', new_metrics_pb2.CUMULATIVE),
                 (False, None, new_metrics_pb2.GAUGE)]
    for cumulative, desc, stream_kind in scenarios:
      _is_cumulative.return_value = cumulative
      m = metrics.Metric('test', description=desc,
                         units=metrics.MetricsDataUnits.SECONDS)
      data_set = new_metrics_pb2.MetricsDataSet()
      m._populate_data_set(data_set, fields={})

      self.assertEqual(stream_kind, data_set.stream_kind)
      self.assertEqual('/infra/test/test', data_set.metric_name)
      self.assertEqual(desc or '', data_set.description)
      self.assertEqual('s', data_set.annotations.unit)

  @mock.patch('infra_libs.ts_mon.common.metrics.Metric._populate_value_new',
              autospec=True)
  @mock.patch('infra_libs.ts_mon.common.metrics.Metric._populate_fields_new',
              autospec=True)
  def test_populate_data(self, _fields, _value):
    m = metrics.Metric('test')
    data_set = new_metrics_pb2.MetricsDataSet()
    m._populate_data(data_set, 100.4, 1000.6, {}, 5)

    self.assertEqual(100, data_set.data[0].start_timestamp.seconds)
    self.assertEqual(1000, data_set.data[0].end_timestamp.seconds)

    _fields.assert_has_calls([mock.call(m, data_set.data[0], {})])
    _value.assert_has_calls([mock.call(m, data_set.data[0], 5)])

  def test_populate_field_descriptor(self):
    data_set_pb = new_metrics_pb2.MetricsDataSet()
    fields = [('a', 1), ('b', True), ('c', 'test')]
    m = metrics.Metric('test')
    m._populate_field_descriptors(data_set_pb, fields)

    field_type = new_metrics_pb2.MetricsDataSet.MetricFieldDescriptor
    self.assertEqual(3, len(data_set_pb.field_descriptor))

    self.assertEqual('a', data_set_pb.field_descriptor[0].name)
    self.assertEqual(field_type.INT64,
                     data_set_pb.field_descriptor[0].field_type)

    self.assertEqual('b', data_set_pb.field_descriptor[1].name)
    self.assertEqual(field_type.BOOL,
                     data_set_pb.field_descriptor[1].field_type)

    self.assertEqual('c', data_set_pb.field_descriptor[2].name)
    self.assertEqual(field_type.STRING,
                     data_set_pb.field_descriptor[2].field_type)

  def test_populate_field_descriptor_error(self):
    data_set_pb = new_metrics_pb2.MetricsDataSet()
    fields = [('a', 1.234)]
    m = metrics.Metric('test')

    with self.assertRaises(errors.MonitoringInvalidFieldTypeError):
      m._populate_field_descriptors(data_set_pb, fields)

  def test_populate_fields(self):
    data = new_metrics_pb2.MetricsData()
    fields = [('a', 1), ('b', True), ('c', 'test')]
    m = metrics.Metric('test')
    m._populate_fields_new(data, fields)

    self.assertEqual(3, len(data.field))

    self.assertEqual('a', data.field[0].name)
    self.assertEqual(1, data.field[0].int64_value)

    self.assertEqual('b', data.field[1].name)
    self.assertTrue(data.field[1].bool_value)

    self.assertEqual('c', data.field[2].name)
    self.assertEqual('test', data.field[2].string_value)

  def test_populate_fields_error(self):
    data = new_metrics_pb2.MetricsData()
    fields = [('a', 1.234)]
    m = metrics.Metric('test')

    with self.assertRaises(errors.MonitoringInvalidFieldTypeError):
      m._populate_fields_new(data, fields)


class StringMetricTest(TestBase):

  def test_populate_value(self):
    pb = metrics_pb2.MetricsData()
    m = metrics.StringMetric('test')
    m._populate_value(pb, 'foo', 1234)
    self.assertEquals(pb.string_value, 'foo')

  def test_populate_value_new(self):
    pb = new_metrics_pb2.MetricsData()
    m = metrics.StringMetric('test')
    m._populate_value_new(pb, 'foo')
    self.assertEquals(pb.string_value, 'foo')

  def test_populate_value_type(self):
    pb = new_metrics_pb2.MetricsDataSet()
    m = metrics.StringMetric('test')
    m._populate_value_type(pb)
    self.assertEquals(pb.value_type, new_metrics_pb2.STRING)

  def test_set(self):
    m = metrics.StringMetric('test')
    m.set('hello world')
    self.assertEquals(m.get(), 'hello world')

  def test_non_string_raises(self):
    m = metrics.StringMetric('test')
    with self.assertRaises(errors.MonitoringInvalidValueTypeError):
      m.set(object())

  def test_is_cumulative(self):
    m = metrics.StringMetric('test')
    self.assertFalse(m.is_cumulative())


class BooleanMetricTest(TestBase):

  def test_populate_value(self):
    pb = metrics_pb2.MetricsData()
    m = metrics.BooleanMetric('test')
    m._populate_value(pb, True, 1234)
    self.assertEquals(pb.boolean_value, True)

  def test_populate_value_new(self):
    pb = new_metrics_pb2.MetricsData()
    m = metrics.BooleanMetric('test')
    m._populate_value_new(pb, True)
    self.assertEquals(pb.bool_value, True)

  def test_populate_value_type(self):
    pb = new_metrics_pb2.MetricsDataSet()
    m = metrics.BooleanMetric('test')
    m._populate_value_type(pb)
    self.assertEquals(pb.value_type, new_metrics_pb2.BOOL)

  def test_set(self):
    m = metrics.BooleanMetric('test')
    m.set(False)
    self.assertEquals(m.get(), False)

  def test_non_bool_raises(self):
    m = metrics.BooleanMetric('test')
    with self.assertRaises(errors.MonitoringInvalidValueTypeError):
      m.set(object())
    with self.assertRaises(errors.MonitoringInvalidValueTypeError):
      m.set('True')
    with self.assertRaises(errors.MonitoringInvalidValueTypeError):
      m.set(123)

  def test_is_cumulative(self):
    m = metrics.BooleanMetric('test')
    self.assertFalse(m.is_cumulative())


class CounterMetricTest(TestBase):

  def test_populate_value(self):
    pb = metrics_pb2.MetricsData()
    m = metrics.CounterMetric('test')
    m._populate_value(pb, 1, 1234)
    self.assertEquals(pb.counter, 1)

  def test_populate_value_new(self):
    pb = new_metrics_pb2.MetricsData()
    m = metrics.CounterMetric('test')
    m._populate_value_new(pb, 1)
    self.assertEquals(pb.int64_value, 1)

  def test_populate_value_type(self):
    pb = new_metrics_pb2.MetricsDataSet()
    m = metrics.CounterMetric('test')
    m._populate_value_type(pb)
    self.assertEquals(pb.value_type, new_metrics_pb2.INT64)

  def test_set(self):
    m = metrics.CounterMetric('test')
    m.set(10)
    self.assertEquals(m.get(), 10)

  def test_increment(self):
    m = metrics.CounterMetric('test')
    m.set(1)
    self.assertEquals(m.get(), 1)
    m.increment()
    self.assertEquals(m.get(), 2)
    m.increment_by(3)
    self.assertAlmostEquals(m.get(), 5)

  def test_decrement_raises(self):
    m = metrics.CounterMetric('test')
    m.set(1)
    with self.assertRaises(errors.MonitoringDecreasingValueError):
      m.set(0)
    with self.assertRaises(errors.MonitoringDecreasingValueError):
      m.increment_by(-1)

  def test_non_int_raises(self):
    m = metrics.CounterMetric('test')
    with self.assertRaises(errors.MonitoringInvalidValueTypeError):
      m.set(object())
    with self.assertRaises(errors.MonitoringInvalidValueTypeError):
      m.set(1.5)
    with self.assertRaises(errors.MonitoringInvalidValueTypeError):
      m.increment_by(1.5)

  def test_multiple_field_values(self):
    m = metrics.CounterMetric('test')
    m.increment({'foo': 'bar'})
    m.increment({'foo': 'baz'})
    m.increment({'foo': 'bar'})
    self.assertIsNone(None)
    self.assertEquals(2, m.get({'foo': 'bar'}))
    self.assertEquals(1, m.get({'foo': 'baz'}))

  def test_override_fields(self):
    m = metrics.CounterMetric('test', fields={'foo': 'bar'})
    m.increment()
    m.increment({'foo': 'baz'})
    self.assertEquals(1, m.get())
    self.assertEquals(1, m.get({'foo': 'bar'}))
    self.assertEquals(1, m.get({'foo': 'baz'}))

  def test_start_timestamp(self):
    t = targets.DeviceTarget('reg', 'role', 'net', 'host')
    m = metrics.CounterMetric('test', fields={'foo': 'bar'})
    m.increment()
    p = metrics_pb2.MetricsCollection()
    m.serialize_to(p, 1234, (), m.get(), t)
    self.assertEquals(1234000000, p.data[0].start_timestamp_us)

  def test_is_cumulative(self):
    m = metrics.CounterMetric('test')
    self.assertTrue(m.is_cumulative())

  def test_get_all(self):
    m = metrics.CounterMetric('test')
    m.increment()
    m.increment({'foo': 'bar'})
    m.increment({'foo': 'baz', 'moo': 'wibble'})
    self.assertEqual([
        ((), 1),
        ((('foo', 'bar'),), 1),
        ((('foo', 'baz'), ('moo', 'wibble')), 1),
    ], sorted(m.get_all()))


class GaugeMetricTest(TestBase):

  def test_populate_value(self):
    pb = metrics_pb2.MetricsData()
    m = metrics.GaugeMetric('test')
    m._populate_value(pb, 1, 1234)
    self.assertEquals(pb.gauge, 1)

  def test_populate_value_new(self):
    pb = new_metrics_pb2.MetricsData()
    m = metrics.GaugeMetric('test')
    m._populate_value_new(pb, 1)
    self.assertEquals(pb.int64_value, 1)

  def test_populate_value_type(self):
    pb = new_metrics_pb2.MetricsDataSet()
    m = metrics.GaugeMetric('test')
    m._populate_value_type(pb)
    self.assertEquals(pb.value_type, new_metrics_pb2.INT64)

  def test_set(self):
    m = metrics.GaugeMetric('test')
    m.set(10)
    self.assertEquals(m.get(), 10)
    m.set(sys.maxint + 1)
    self.assertEquals(m.get(), sys.maxint + 1)

  def test_non_int_raises(self):
    m = metrics.GaugeMetric('test')
    with self.assertRaises(errors.MonitoringInvalidValueTypeError):
      m.set(object())

  def test_is_cumulative(self):
    m = metrics.GaugeMetric('test')
    self.assertFalse(m.is_cumulative())


class CumulativeMetricTest(TestBase):

  def test_populate_value(self):
    pb = metrics_pb2.MetricsData()
    m = metrics.CumulativeMetric('test')
    m._populate_value(pb, 1.618, 1234)
    self.assertAlmostEquals(pb.cumulative_double_value, 1.618)

  def test_populate_value_new(self):
    pb = new_metrics_pb2.MetricsData()
    m = metrics.CumulativeMetric('test')
    m._populate_value_new(pb, 1.618)
    self.assertAlmostEquals(pb.double_value, 1.618)

  def test_populate_value_type(self):
    pb = new_metrics_pb2.MetricsDataSet()
    m = metrics.CumulativeMetric('test')
    m._populate_value_type(pb)
    self.assertEqual(pb.value_type, new_metrics_pb2.DOUBLE)

  def test_set(self):
    m = metrics.CumulativeMetric('test')
    m.set(3.14)
    self.assertAlmostEquals(m.get(), 3.14)

  def test_decrement_raises(self):
    m = metrics.CumulativeMetric('test')
    m.set(3.14)
    with self.assertRaises(errors.MonitoringDecreasingValueError):
      m.set(0)
    with self.assertRaises(errors.MonitoringDecreasingValueError):
      m.increment_by(-1)

  def test_non_number_raises(self):
    m = metrics.CumulativeMetric('test')
    with self.assertRaises(errors.MonitoringInvalidValueTypeError):
      m.set(object())

  def test_start_timestamp(self):
    t = targets.DeviceTarget('reg', 'role', 'net', 'host')
    m = metrics.CumulativeMetric('test', fields={'foo': 'bar'})
    m.set(3.14)
    p = metrics_pb2.MetricsCollection()
    m.serialize_to(p, 1234, (), m.get(), t)
    self.assertEquals(1234000000, p.data[0].start_timestamp_us)

  def test_is_cumulative(self):
    m = metrics.CumulativeMetric('test')
    self.assertTrue(m.is_cumulative())


class FloatMetricTest(TestBase):

  def test_populate_value(self):
    pb = metrics_pb2.MetricsData()
    m = metrics.FloatMetric('test')
    m._populate_value(pb, 1.618, 1234)
    self.assertEquals(pb.noncumulative_double_value, 1.618)

  def test_populate_value_new(self):
    pb = new_metrics_pb2.MetricsData()
    m = metrics.FloatMetric('test')
    m._populate_value_new(pb, 1.618)
    self.assertEquals(pb.double_value, 1.618)

  def test_populate_value_type(self):
    pb = new_metrics_pb2.MetricsDataSet()
    m = metrics.FloatMetric('test')
    m._populate_value_type(pb)
    self.assertEquals(pb.value_type, new_metrics_pb2.DOUBLE)

  def test_set(self):
    m = metrics.FloatMetric('test')
    m.set(3.14)
    self.assertEquals(m.get(), 3.14)

  def test_non_number_raises(self):
    m = metrics.FloatMetric('test')
    with self.assertRaises(errors.MonitoringInvalidValueTypeError):
      m.set(object())

  def test_is_cumulative(self):
    m = metrics.FloatMetric('test')
    self.assertFalse(m.is_cumulative())


class RunningZeroGeneratorTest(TestBase):

  def assertZeroes(self, expected, sequence):
    self.assertEquals(expected,
        list(metrics.DistributionMetric._running_zero_generator(sequence)))

  def test_running_zeroes(self):
    self.assertZeroes([1, -1, 1], [1, 0, 1])
    self.assertZeroes([1, -2, 1], [1, 0, 0, 1])
    self.assertZeroes([1, -3, 1], [1, 0, 0, 0, 1])
    self.assertZeroes([1, -1, 1, -1, 2], [1, 0, 1, 0, 2])
    self.assertZeroes([1, -1, 1, -2, 2], [1, 0, 1, 0, 0, 2])
    self.assertZeroes([1, -2, 1, -2, 2], [1, 0, 0, 1, 0, 0, 2])

  def test_leading_zeroes(self):
    self.assertZeroes([-1, 1], [0, 1])
    self.assertZeroes([-2, 1], [0, 0, 1])
    self.assertZeroes([-3, 1], [0, 0, 0, 1])

  def test_trailing_zeroes(self):
    self.assertZeroes([1], [1])
    self.assertZeroes([1], [1, 0])
    self.assertZeroes([1], [1, 0, 0])
    self.assertZeroes([], [])
    self.assertZeroes([], [0])
    self.assertZeroes([], [0, 0])


class DistributionMetricTest(TestBase):

  def test_populate_canonical(self):
    pb = metrics_pb2.MetricsData()
    m = metrics.DistributionMetric('test')
    m._populate_value(pb,
        distribution.Distribution(distribution.GeometricBucketer()),
        1234)
    self.assertEquals(pb.distribution.spec_type,
        metrics_pb2.PrecomputedDistribution.CANONICAL_POWERS_OF_10_P_0_2)

    m._populate_value(pb,
        distribution.Distribution(distribution.GeometricBucketer(2)),
        1234)
    self.assertEquals(pb.distribution.spec_type,
        metrics_pb2.PrecomputedDistribution.CANONICAL_POWERS_OF_2)

    m._populate_value(pb,
        distribution.Distribution(distribution.GeometricBucketer(10)),
        1234)
    self.assertEquals(pb.distribution.spec_type,
        metrics_pb2.PrecomputedDistribution.CANONICAL_POWERS_OF_10)

  def test_populate_custom(self):
    pb = metrics_pb2.MetricsData()
    m = metrics.DistributionMetric('test')
    m._populate_value(pb,
        distribution.Distribution(distribution.GeometricBucketer(4)),
        1234)
    self.assertEquals(pb.distribution.spec_type,
        metrics_pb2.PrecomputedDistribution.CUSTOM_PARAMETERIZED)
    self.assertEquals(0, pb.distribution.width)
    self.assertEquals(4, pb.distribution.growth_factor)
    self.assertEquals(100, pb.distribution.num_buckets)

    m._populate_value(pb,
        distribution.Distribution(distribution.FixedWidthBucketer(10)),
        1234)
    self.assertEquals(pb.distribution.spec_type,
        metrics_pb2.PrecomputedDistribution.CUSTOM_PARAMETERIZED)
    self.assertEquals(10, pb.distribution.width)
    self.assertEquals(0, pb.distribution.growth_factor)
    self.assertEquals(100, pb.distribution.num_buckets)

  def test_populate_buckets(self):
    pb = metrics_pb2.MetricsData()
    m = metrics.DistributionMetric('test')
    d = distribution.Distribution(
        distribution.FixedWidthBucketer(10))
    d.add(5)
    d.add(15)
    d.add(35)
    d.add(65)

    m._populate_value(pb, d, 1234)
    self.assertEquals([1, 1, -1, 1, -2, 1], pb.distribution.bucket)
    self.assertEquals(0, pb.distribution.underflow)
    self.assertEquals(0, pb.distribution.overflow)
    self.assertEquals(30, pb.distribution.mean)

    pb = metrics_pb2.MetricsData()
    d = distribution.Distribution(
        distribution.FixedWidthBucketer(10, num_finite_buckets=1))
    d.add(5)
    d.add(15)
    d.add(25)

    m._populate_value(pb, d, 1234)
    self.assertEquals([1], pb.distribution.bucket)
    self.assertEquals(0, pb.distribution.underflow)
    self.assertEquals(2, pb.distribution.overflow)
    self.assertEquals(15, pb.distribution.mean)

  def test_populate_buckets_last_zero(self):
    pb = metrics_pb2.MetricsData()
    m = metrics.DistributionMetric('test')
    d = distribution.Distribution(
        distribution.FixedWidthBucketer(10, num_finite_buckets=10))
    d.add(5)
    d.add(105)

    m._populate_value(pb, d, 1234)
    self.assertEquals([1], pb.distribution.bucket)
    self.assertEquals(1, pb.distribution.overflow)

  def test_populate_buckets_underflow(self):
    pb = metrics_pb2.MetricsData()
    m = metrics.DistributionMetric('test')
    d = distribution.Distribution(
        distribution.FixedWidthBucketer(10, num_finite_buckets=10))
    d.add(-5)
    d.add(-1000000)

    m._populate_value(pb, d, 1234)
    self.assertEquals([], pb.distribution.bucket)
    self.assertEquals(2, pb.distribution.underflow)
    self.assertEquals(0, pb.distribution.overflow)
    self.assertEquals(-500002.5, pb.distribution.mean)

  def test_populate_is_cumulative(self):
    pb = metrics_pb2.MetricsData()
    d = distribution.Distribution(
        distribution.FixedWidthBucketer(10, num_finite_buckets=10))
    m = metrics.CumulativeDistributionMetric('test')

    m._populate_value(pb, d, 1234)
    self.assertTrue(pb.distribution.is_cumulative)

    m = metrics.NonCumulativeDistributionMetric('test2')

    m._populate_value(pb, d, 1234)
    self.assertFalse(pb.distribution.is_cumulative)

  def test_populate_fixed_width_buckets_new(self):
    pb = new_metrics_pb2.MetricsData()
    dist = distribution.Distribution(
        distribution.FixedWidthBucketer(width=9, num_finite_buckets=10))
    dist.add(1)
    dist.add(10)
    dist.add(100)
    dist.add(1000)
    m = metrics.DistributionMetric('test')
    m._populate_value_new(pb, dist)

    self.assertAlmostEqual(277.75, pb.distribution_value.mean)
    self.assertEqual(9, pb.distribution_value.linear_buckets.width)
    self.assertEqual(0, pb.distribution_value.linear_buckets.offset)
    self.assertEqual(10,
                     pb.distribution_value.linear_buckets.num_finite_buckets)
    self.assertEqual([0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 2],
                     list(pb.distribution_value.bucket_count))

  def test_populate_geometric_buckets_new(self):
    pb = new_metrics_pb2.MetricsData()
    dist = distribution.Distribution(
        distribution.GeometricBucketer(growth_factor=10**0.2,
                                       num_finite_buckets=10))
    dist.add(1)
    dist.add(10)
    dist.add(100)
    dist.add(1000)
    m = metrics.DistributionMetric('test')
    m._populate_value_new(pb, dist)

    self.assertAlmostEqual(277.75, pb.distribution_value.mean)
    self.assertAlmostEqual(1.0, pb.distribution_value.exponential_buckets.scale)
    self.assertAlmostEqual(
         10**0.2, pb.distribution_value.exponential_buckets.growth_factor)
    self.assertEqual(
        10, pb.distribution_value.exponential_buckets.num_finite_buckets)
    self.assertEqual([0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 2],
                     list(pb.distribution_value.bucket_count))

  def test_populate_value_type(self):
    pb = new_metrics_pb2.MetricsDataSet()
    m = metrics.CumulativeDistributionMetric('test')
    m._populate_value_type(pb)
    self.assertEqual(pb.value_type, new_metrics_pb2.DISTRIBUTION)

  def test_add(self):
    m = metrics.DistributionMetric('test')
    m.add(1)
    m.add(10)
    m.add(100)
    self.assertEquals({2: 1, 6: 1, 11: 1}, m.get().buckets)
    self.assertEquals(111, m.get().sum)
    self.assertEquals(3, m.get().count)

  def test_add_custom_bucketer(self):
    m = metrics.DistributionMetric('test',
        bucketer=distribution.FixedWidthBucketer(10))
    m.add(1)
    m.add(10)
    m.add(100)
    self.assertEquals({1: 1, 2: 1, 11: 1}, m.get().buckets)
    self.assertEquals(111, m.get().sum)
    self.assertEquals(3, m.get().count)

  def test_set(self):
    d = distribution.Distribution(
        distribution.FixedWidthBucketer(10, num_finite_buckets=10))
    d.add(1)
    d.add(10)
    d.add(100)

    m = metrics.CumulativeDistributionMetric('test')
    with self.assertRaises(TypeError):
      m.set(d)

    m = metrics.NonCumulativeDistributionMetric('test2')
    m.set(d)
    self.assertEquals(d, m.get())

    with self.assertRaises(errors.MonitoringInvalidValueTypeError):
      m.set(1)
    with self.assertRaises(errors.MonitoringInvalidValueTypeError):
      m.set('foo')

  def test_start_timestamp(self):
    t = targets.DeviceTarget('reg', 'role', 'net', 'host')
    m = metrics.CumulativeDistributionMetric('test')
    m.add(1)
    m.add(5)
    m.add(25)
    p = metrics_pb2.MetricsCollection()
    m.serialize_to(p, 1234, (), m.get(), t)
    self.assertEquals(1234000000, p.data[0].start_timestamp_us)

  def test_is_cumulative(self):
    cd = metrics.CumulativeDistributionMetric('test')
    ncd = metrics.NonCumulativeDistributionMetric('test2')
    self.assertTrue(cd.is_cumulative())
    self.assertFalse(ncd.is_cumulative())
