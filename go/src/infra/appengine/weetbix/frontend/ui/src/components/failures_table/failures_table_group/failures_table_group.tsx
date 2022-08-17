// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// import { nanoid } from 'nanoid';
import {
  Fragment,
  ReactNode,
} from 'react';

import {
  FailureGroup,
  GroupKey,
  VariantGroup,
} from '../../../tools/failures_tools';
import FailuresTableRows from './failures_table_rows/failures_table_rows';

const renderGroup = (
    project: string,
    parentKeys: GroupKey[],
    group: FailureGroup,
    variantGroups: VariantGroup[],
): ReactNode => {
  return (
    <Fragment>
      <FailuresTableRows
        project={project}
        parentKeys={parentKeys}
        group={group}
        variantGroups={variantGroups}>
        {
          group.children.map((childGroup) => (
            <Fragment key={childGroup.id}>
              {renderGroup(project, [...parentKeys, group.key], childGroup, variantGroups)}
            </Fragment>
          ))
        }
      </FailuresTableRows>
    </Fragment>
  );
};

interface Props {
  project: string;
  parentKeys?: GroupKey[];
  group: FailureGroup;
  variantGroups: VariantGroup[];
}

const FailuresTableGroup = ({
  project,
  parentKeys = [],
  group,
  variantGroups,
}: Props) => {
  return (
    <>{renderGroup(project, parentKeys, group, variantGroups)}</>
  );
};

export default FailuresTableGroup;
