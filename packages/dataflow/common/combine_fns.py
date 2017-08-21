# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file

import apache_beam as beam


class ConvertToCSV(beam.CombineFn):
  """Convert elements to CSV format to be written out

  Transform for writing elements out in a CSV format. Can process elements
  of type dictonary or list. This transform only supports consistent elements,
  meaning dictionaries must all have the same keys and lists must have the
  same length and order.
  """
  def create_accumulator(self):
    return []

  def iterable(self, obj):
    # Sort ensures the fields of the element are written in the
    # same order if it is a dictionary.
    if isinstance(obj, dict):
      keys = obj.keys()
      return sorted(keys)
    else:
      return range(len(obj))

  def add_input(self, accumulator, element):
    element_string = []
    for i in self.iterable(element):
      element_string.append(str(element[i]))
    accumulator.append(','.join(element_string) + '\n')
    return accumulator

  def merge_accumulators(self, accumulators):
    return sum(accumulators, [])

  def extract_output(self, accumulator):
    return ''.join(accumulator)
