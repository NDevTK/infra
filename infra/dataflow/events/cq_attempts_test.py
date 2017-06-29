import unittest
import cq_attempts as job

class TestCQAttemptAccumulator(unittest.TestCase):
  def setUp(self):
    self.attempt_start_usec = 1493833887566000
    self.attempt_start_msec = self.attempt_start_usec / 1000
    self.patch_start = 1493833887688

    accumulator = job.CQAttempt()
    accumulator.attempt_start_msec = self.attempt_start_msec
    accumulator.first_start_msec = self.patch_start
    accumulator.last_start_msec = self.patch_start
    self.basic_accumulator = accumulator

    self.combFn = job.CombineEventsToAttempt()

  def test_add_first_start(self):
    accumulator = job.CQAttempt()
    rows = [{
        'attempt_start_usec': self.attempt_start_usec,
        'timestamp_millis': self.patch_start,
        'action': job.ACTION_PATCH_START,
    }]
    accumulator = self.combFn.add_input(accumulator, rows)
    self.assertEqual(accumulator.first_start_msec, self.patch_start)
    self.assertEqual(accumulator.last_start_msec, self.patch_start)

  def test_add_null_start(self):
    accumulator = self.basic_accumulator
    rows = [{
        'attempt_start_usec': self.attempt_start_usec,
        'timestamp_millis': None,
        'action': job.ACTION_PATCH_START,
    }]
    accumulator = self.combFn.add_input(accumulator, rows)
    self.assertEqual(accumulator.first_start_msec, self.patch_start)
    self.assertEqual(accumulator.last_start_msec, self.patch_start)

  def test_add_earlier_start(self):
    accumulator = self.basic_accumulator
    earlier_patch_start = self.patch_start - 1
    rows = [{
        'attempt_start_usec': self.attempt_start_usec,
        'timestamp_millis': earlier_patch_start,
        'action': job.ACTION_PATCH_START,
    }]
    accumulator = self.combFn.add_input(accumulator, rows)
    self.assertEqual(accumulator.first_start_msec, earlier_patch_start)
    self.assertEqual(accumulator.last_start_msec, self.patch_start)

  def test_add_later_start(self):
    accumulator = self.basic_accumulator
    later_patch_start = self.patch_start + 1
    rows = [{
        'attempt_start_usec': self.attempt_start_usec,
        'timestamp_millis': later_patch_start,
        'action': job.ACTION_PATCH_START,
    }]
    accumulator = self.combFn.add_input(accumulator, rows)
    self.assertEqual(accumulator.first_start_msec, self.patch_start)
    self.assertEqual(accumulator.last_start_msec, later_patch_start)

  def test_merge_null_start(self):
    accumulator = self.basic_accumulator
    another = job.CQAttempt()
    merged = self.combFn.merge_accumulators([accumulator, another])
    self.assertEqual(merged.first_start_msec, self.patch_start)
    self.assertEqual(merged.last_start_msec, self.patch_start)

  def test_merge_earlier_start(self):
    accumulator = self.basic_accumulator
    another = self.basic_accumulator
    earlier_patch_start = self.patch_start - 1
    another.first_start_msec = earlier_patch_start
    merged = self.combFn.merge_accumulators([accumulator, another])
    self.assertEqual(merged.first_start_msec, earlier_patch_start)
    self.assertEqual(merged.last_start_msec, self.patch_start)

  def test_merge_later_start(self):
    accumulator = self.basic_accumulator
    another = self.basic_accumulator
    later_patch_start = self.patch_start + 1
    another.last_start_msec = later_patch_start
    merged = self.combFn.merge_accumulators([accumulator, another])
    self.assertEqual(merged.first_start_msec, self.patch_start)
    self.assertEqual(merged.last_start_msec, later_patch_start)

if __name__ == '__main__':
  unittest.main()
