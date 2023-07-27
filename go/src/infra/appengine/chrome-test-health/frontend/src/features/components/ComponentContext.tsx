// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { createContext, useEffect, useState } from 'react';
import { listComponents } from '../../api/resources';

type ComponentContextProviderProps = {
    children: React.ReactNode,
    components?: string[],
  }

export interface ComponentContextValue {
    components: string[],
    allComponents: string[],
    api: Api
}

export interface Api {
    // Component navigation
    updateComponents: (component: string[]) => void,
}

export const ComponentContext = createContext<ComponentContextValue>(
    {
      components: [],
      allComponents: [],
      api: {
        updateComponents: () => {/**/},
      },
    },
);

export const ComponentContextProvider = (props: ComponentContextProviderProps) => {
  const [allComponents, setAllComponents] = useState<string[]>(['Blink']);
  const [components, setComponents] = useState<string[]>(props.components || ['Blink']);

  function loadComponents() {
    listComponents().then((resp) => {
      setAllComponents(resp.components);
    }).catch((err) => {
      throw err;
    });
  }

  useEffect(() => {
    // On mount, populate the components.
    loadComponents();
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const api: Api = {
    updateComponents: (newComponents: string[]) => {
      setComponents(newComponents);
    },
  };

  return (
    <ComponentContext.Provider value={{ components, allComponents, api }}>
      { props.children }
    </ComponentContext.Provider>
  );
};
