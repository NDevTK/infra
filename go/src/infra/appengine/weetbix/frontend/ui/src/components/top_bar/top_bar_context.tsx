// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import {
  createContext,
  Dispatch,
  SetStateAction,
  useState,
} from 'react';

type AnchorElNav = null | HTMLElement;

export interface TopBarContextState {
  anchorElNav: AnchorElNav;
  setAnchorElNav: Dispatch<SetStateAction<AnchorElNav>>
}

export const TopBarContext = createContext<TopBarContextState>({
  anchorElNav: null,
  // eslint-disable-next-line @typescript-eslint/no-empty-function
  setAnchorElNav: () => {},
});

interface Props {
  children: React.ReactNode;
}

export const TopBarContextProvider = ({ children }: Props) => {
  const [anchorElNav, setAnchorElNav] = useState<AnchorElNav>(null);
  return (
    <TopBarContext.Provider value={{ anchorElNav, setAnchorElNav }}>
      {children}
    </TopBarContext.Provider>
  );
};
