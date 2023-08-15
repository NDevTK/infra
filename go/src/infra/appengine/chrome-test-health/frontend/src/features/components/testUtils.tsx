// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { render } from '@testing-library/react';
import { ComponentContext } from './ComponentContext';

interface OptionalContext {
  components?: string[],
  allComponents?: string[],
  api?: {
    updateComponents?: (components: string[]) => void,
  },
}

export function renderWithComponents(ui: React.ReactElement, opts: OptionalContext = {}) {
  return render((
    <ComponentContext.Provider value={{
      components: opts.components || [],
      allComponents: opts.allComponents || [],
      api: {
        updateComponents: opts?.api?.updateComponents || (() => {/**/}),
      },
    }}>
      {ui}
    </ComponentContext.Provider>
  ));
}
