import logging
import time

import apache_beam as beam

ACTION_PATCH_START = 'PATCH_START'

class CQAttempt(object):
  def __init__(self):
    self.attempt_start_msec = None
    self.first_start_msec = None
    self.last_start_msec = None

  def update_first_start(self, new_timestamp):
    if new_timestamp is None:
      return
    if self.first_start_msec is None or new_timestamp < self.first_start_msec:
      self.first_start_msec = new_timestamp

  def update_last_start(self, new_timestamp):
    if new_timestamp is None:
      return
    if self.last_start_msec is None or new_timestamp > self.first_start_msec:
      self.last_start_msec = new_timestamp

class CombineEventsToAttempt(beam.CombineFn):
  def create_accumulator(self):
    return CQAttempt()

  def add_input(self, accumulator, i):
    for row in i:
      if row.get('attempt_start_usec') is None:
        logging.warn('recieved row with null attempt_start_usec: %s', row)
        continue

      timestamp = row.get('timestamp_millis')
      if timestamp is None:
        logging.warn('recieved raw with null timestamp: %s', row)
        continue

      attempt_start_msec = row.get('attempt_start_usec') / 1000
      if accumulator.attempt_start_msec is None:
        accumulator.attempt_start_msec = attempt_start_msec
      action = row.get('action')

      if action == ACTION_PATCH_START:
        accumulator.update_first_start(timestamp)
        accumulator.update_last_start(timestamp)

    return accumulator

  def merge_accumulators(self, accumulators):
    if len(accumulators) == 1:
      return accumulators[0]
    merged = accumulators[0]
    for a in accumulators[1:]:
      merged.update_first_start(a.first_start_msec)
      merged.update_last_start(a.last_start_msec)
    return merged

  def extract_output(self, a):
    return a.__dict__

def main():
  one_day_ago_usec = time.time() * 1000000 - 24 * 60 * 60 * 1000000
  q = "SELECT timestamp_millis, action, attempt_start_usec \
       FROM `chrome-infra-events.raw_events.cq` \
       WHERE attempt_start_usec > %d" % one_day_ago_usec
  p = beam.Pipeline(argv=['--project', 'chrome-infra-events'])
  _ = (p
   | beam.Read(beam.io.BigQuerySource(query=q,
                                      flatten_results=False,
                                      use_standard_sql=True))
   | beam.Map(lambda e: (e['attempt_start_usec'], e))
   | beam.GroupByKey()
   | beam.CombinePerKey(CombineEventsToAttempt())
   | beam.Map(lambda (k, v): v)
   | beam.io.Write(beam.io.BigQuerySink('cq_attempts', dataset='aggregated',
                                        project='chrome-infra-events')))
  p.run()

if __name__ == '__main__':
  main()
