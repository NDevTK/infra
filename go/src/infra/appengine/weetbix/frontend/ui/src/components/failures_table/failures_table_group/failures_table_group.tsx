// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// import { nanoid } from 'nanoid';
import React, {
    Fragment,
    ReactNode
} from 'react';

import {
    FailureGroup,
    VariantGroup
} from '../../../tools/failures_tools';
import FailuresTableRows from './failures_table_rows/failures_table_rows';

const renderGroup = (
    group: FailureGroup,
    variantGroups: VariantGroup[]
): ReactNode => {

    return (
        <Fragment>
            <FailuresTableRows
                group={group}
                variantGroups={variantGroups}
            >
                {
                    group.children.map(childGroup => (
                        <Fragment key={childGroup.id}>
                            {renderGroup(childGroup, variantGroups)}
                        </Fragment>
                    ))
                }
            </FailuresTableRows>
        </Fragment>
    );
};

interface Props {
    group: FailureGroup;
    variantGroups: VariantGroup[];
};

const FailuresTableGroup = ({
    group,
    variantGroups
}: Props) => {
    return (
        <>{renderGroup(group, variantGroups)}</>
    );
};

export default FailuresTableGroup;