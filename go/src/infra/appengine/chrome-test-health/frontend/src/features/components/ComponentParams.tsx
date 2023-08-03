// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { useCallback, useContext, useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';
import { ComponentContext } from './ComponentContext';

export const COMPONENT = 'comp';

function ComponentParams() {
  const { components } = useContext(ComponentContext);

  const [, setSearchParams] = useSearchParams();

  const updateParams = useCallback((search: URLSearchParams) => {
    search.delete(COMPONENT);
    components.forEach((c) => search.append(COMPONENT, c));
    localStorage.setItem(COMPONENT, components.join(','));
    return search;
  }, [components]);

  useEffect(() => {
    setSearchParams(updateParams);
  }, [components, setSearchParams, updateParams]);
  return (<></>);
}

export default ComponentParams;
