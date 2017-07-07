class BigQueryObject(object):
  def get_bigquery_attributes(self):
    raise NotImplementedError

  def as_bigquery_row(self):
    row = {}
    for attr in self.get_bigquery_attributes():
      row[attr] = self.__dict__.get(attr)
    return row

class CQAttempt(BigQueryObject):
  def __init__(self):
    self.attempt_start_msec = None
    self.first_start_msec = None
    self.last_start_msec = None

  def get_bigquery_attributes(self):
    return [
        'attempt_start_msec',
        'first_start_msec',
        'last_start_msec',
    ]

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
