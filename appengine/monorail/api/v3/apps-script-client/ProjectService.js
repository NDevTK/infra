// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

/* eslint-disable no-unused-vars */

const MAX_COMPONENT_PAGE_SIZE = 100;

/**
 * Creates a ComponentDef.
 * @param {string} projectName The resource name of the parent project.
 * @param {string} value The name of the component
 *     e.g. "Triage" or "Triage>Security".
 * @param {string=} docstring Short description of the ComponentDef.
 * @param {Array<string>=} admins Array of User resource names to set as admins.
 * @param {Array<string>=} ccs Array of User resources names to set as auto-ccs.
 * @param {Array<string>=} labels Array of labels.
 * @return {ComponentDef}
 */
function createComponentDef(
    projectName, value, docstring, admins, ccs, labels) {
  const componentDef = {
    'value': value,
    'docstring': docstring,
  };
  if (admins) {
    componentDef['admins'] = admins;
  }
  if (ccs) {
    componentDef['ccs'] = ccs;
  }
  if (labels) {
    componentDef['labels'] = labels;
  }
  const message = {
    'parent': projectName,
    'componentDef': componentDef,
  };
  const url = URL + 'monorail.v3.Projects/CreateComponentDef';
  return run_(url, message);
}

/**
 * Deletes a ComponentDef.
 * @param {string} componentName Resource name of the ComponentDef to delete.
 * @return {EmptyProto}
 */
function deleteComponentDef(componentName) {
  const message = {
    'name': componentName,
  };
  const url = URL + 'monorail.v3.Projects/DeleteComponentDef';
  return run_(url, message);
}

/**
 * Lists all ComponentDefs for a project. Automatically traverses through pages
 * to get all components.
 * @param {string=} projectName Resource name of the project to fetch components for.
 * @param {string=} includeDeprecated Where to include deprecated components.
 * @return {Array<ComponentDef>}
 */
function listComponentDefs(projectName = 'projects/chromium', includeDeprecated=false) {
  const url = URL + 'monorail.v3.Projects/ListComponentDefs';
  let components = [];

  let response;
  do {
    const message = {
      'parent': projectName,
      'pageSize': MAX_COMPONENT_PAGE_SIZE,
      'pageToken': response ? response.nextPageToken : undefined,
    };
    response = run_(url, message);
    components = [...components, ...response.componentDefs];
  } while (response.nextPageToken);

  if (!includeDeprecated) {
    components = components.filter((c) => c.state === 'ACTIVE');
  }
  return components;
}
