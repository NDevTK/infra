// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import React from 'react';
import { makeStyles } from '@material-ui/styles';
import { RoleSelection } from './RoleSelection/RoleSelection';
import {IssueWizardUserGroup} from '../IssueWizardTypes';

const useStyles = makeStyles({
  flex: {
    display: 'flex',
    justifyContent: 'space-between',
  }
});

const getUserGroupSelectors = (
  value: string,
  onSelectorClick:
    (selector: string) =>
      (event: React.MouseEvent<HTMLElement>) => any) => {
  const selectors = new Array();
  Object.values(IssueWizardUserGroup).forEach((userGroup) => {
    selectors.push(
        <RoleSelection
          checked={value === userGroup.name}
          handleOnClick={onSelectorClick(userGroup.name)}
          value={userGroup.name}
          description={userGroup.description}
          inputProps={{ 'aria-label': userGroup.name }}
        />
      );
  });
  return selectors;
}
/**
 * RadioDescription contains a set of radio buttons and descriptions (RoleSelection)
 * to be chosen from in the landing step of the Issue Wizard.
 *
 * @returns React.ReactElement
 */
export const RadioDescription = ({ value, setValue }: { value: string, setValue: Function }): React.ReactElement => {
  const classes = useStyles();

  const handleRoleSelectionClick = (userGroup: string) =>
     (event: React.MouseEvent<HTMLElement>) => setValue(userGroup);

  const userGroupsSelectors = getUserGroupSelectors(value, handleRoleSelectionClick);

  return (
    <div className={classes.flex}>
      {userGroupsSelectors}
    </div>
  );
}
