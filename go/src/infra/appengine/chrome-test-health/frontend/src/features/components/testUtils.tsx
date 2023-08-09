// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { render } from '@testing-library/react';
import { ComponentContext } from './ComponentContext';

export function renderWithComponents(ui: React.ReactElement, components: string[] = []) {
  return render((
    <ComponentContext.Provider value={{
      components,
      allComponents: [],
      api: { updateComponents: () => {/**/} },
    }}>
      {ui}
    </ComponentContext.Provider>
  ));
}
