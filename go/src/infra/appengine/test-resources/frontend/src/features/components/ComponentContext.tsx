// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { createContext, useEffect, useState } from 'react';
import { listComponents } from '../../api/resources';

type ComponentContextProviderProps = {
    children: React.ReactNode,
    component?: string,
  }

export interface ComponentContextValue {
    component: string,
    allComponents: string[],
    api: Api
}

export interface Api {
    // Component navigation
    updateComponent: (component: string) => void,
}

export const ComponentContext = createContext<ComponentContextValue>(
    {
      component: '',
      allComponents: [],
      api: {
        updateComponent: () => {/**/},
      },
    },
);

export const ComponentContextProvider = (props: ComponentContextProviderProps) => {
  const [component, setComponent] = useState('Blink');
  const [allComponents, setAllComponents] = useState<string[]>([]);

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
    updateComponent: (newComponent: string) => {
      setComponent(newComponent);
    },
  };

  return (
    <ComponentContext.Provider value={{ component, allComponents, api }}>
      { props.children }
    </ComponentContext.Provider>
  );
};
