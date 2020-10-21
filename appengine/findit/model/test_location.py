from google.appengine.ext import ndb


class TestLocation(ndb.Model):
  """The location of a test in the source tree"""
  file_path = ndb.StringProperty(required=True)
  line_number = ndb.IntegerProperty(required=True)
