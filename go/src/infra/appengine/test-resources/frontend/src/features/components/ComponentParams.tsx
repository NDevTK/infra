// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { useCallback, useContext, useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';
import { ComponentContext } from './ComponentContext';

export const COMPONENT = 'comp';

function ComponentParams() {
  const { components } = useContext(ComponentContext);

  const [search, setSearchParams] = useSearchParams();

  const updateParams = useCallback(() => {
    search.delete(COMPONENT);
    components.forEach((c) => search.append(COMPONENT, c));
    setSearchParams(search);
  }, [search, setSearchParams, components]);

  useEffect(() => {
    updateParams();
  }, [updateParams]);

  return (<></>);
}

export default ComponentParams;
