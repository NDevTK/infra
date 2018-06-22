"""Cron job that adds all new user commmits to the SQL database daily."""

import logging
import settings
import time
import cgi
import csv
import logging
import time
import webapp2
import cloudstorage
import json

from googleapiclient import discovery
from googleapiclient import errors
from google.appengine.api import urlfetch
from framework import jsonfeed
from google.appengine.api import app_identity
from oauth2client.client import GoogleCredentials
import webapp2

class GetCommitsCron(webapp2.RequestHandler):

  """Fetches commit data from Gitiles and adds it to the CloudSQL database
  """
  def get(self):
      url = 'https://gerrit.googlesource.com/gerrit/+log/?format=JSON'
      try:
          result = urlfetch.fetch(url)
          if result.status_code == 200:
              self.response.write(result.content)
          else:
              self.response.status_code = result.status_code
      except urlfetch.Error:
          logging.exception('Caught exception fetching url')
