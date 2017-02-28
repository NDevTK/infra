# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Convert directory-component mapping from OWNERS files into Predator
required format.
"""

import os
import json
import urllib2
from collections import defaultdict


DEFAULT_MAPPING_URL = \
    'https://storage.googleapis.com/chromium-owners/component_map.json'


def ConvertComponentMappingFormat(url=DEFAULT_MAPPING_URL):
  """Convert component mapping from owners files into componnet classifier
  config required format.

  The main purpose is to get the latest component/team information from
  OWNERS files and convert it into format in the component classifier
  config in Predator internal config page.

  Args:
    url: url link to the latest component_map from OWNERS files

  Returns:
    a list where each element is a dict of form {
    'component': component name,
    'function': function,
    'dir': a list of path maps to this component,
    'team': the team mailing list responsible to triage this component
    }
  """
  mappings_file = json.load(urllib2.urlopen(url))
  component_dict = defaultdict(dict)
  for dir_name, component in mappings_file['dir-to-component'].items():
    if component_dict[component].get('dir'):
      component_dict[component]['dir'].append(dir_name)
    else:
      component_dict[component]['dir'] = [dir_name]
    component_dict[component]['team'] = \
        mappings_file['component-to-team'].get(component)

  component_list = []
  for component, value in component_dict.items():
    component_list.append({
        'component': component,
        'dir': value['dir'],
        'team': value['team'],
        'function': value.get('function')})

  return component_list
