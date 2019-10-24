import os
import sys

# Append paths so that dependencies would work.
_FINDIT_DIR = os.path.join(
    os.path.dirname(__file__), os.path.pardir, os.path.pardir)
_THIRD_PARTY_DIR = os.path.join(
    os.path.dirname(__file__), os.path.pardir, os.path.pardir, 'third_party')
_FIRST_PARTY_DIR = os.path.join(
    os.path.dirname(__file__), os.path.pardir, os.path.pardir, 'first_party')
sys.path.insert(0, _FINDIT_DIR)
sys.path.insert(0, _THIRD_PARTY_DIR)
sys.path.insert(0, _FIRST_PARTY_DIR)

# Activate script as findit prod.
from local_libs import remote_api
remote_api.EnableRemoteApi(app_id='findit-for-me-staging')

# Add imports below.
import datetime
import textwrap

from handlers.code_coverage import ProcessCodeCoverageData
x = ProcessCodeCoverageData()
x._processCodeCoverageData(8898859569891525600)
