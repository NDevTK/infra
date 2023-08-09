// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { createContext, useContext, useEffect, useState } from 'react';
import { listComponents } from '../../api/resources';
import { AuthContext } from '../auth/AuthContext';

export const URL_COMPONENT = 'c';

export function updateComponentsUrl(components: string[], search: URLSearchParams) {
  search.delete(URL_COMPONENT);
  components.forEach((c) => search.append(URL_COMPONENT, c));
  // Updating it here isn't great, but not sure there's a cleaner way to do this
  localStorage.setItem(URL_COMPONENT, components.join(','));
}

type ComponentContextProviderProps = {
  children: React.ReactNode,
  components: string[],
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
  const { auth } = useContext(AuthContext);
  const [allComponents, setAllComponents] = useState<string[]>(props.components);
  const [components, setComponents] = useState<string[]>(props.components);

  function loadComponents() {
    if (auth == undefined) {
      return;
    }
    listComponents(auth).then((resp) => {
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
