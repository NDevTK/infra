// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { Outlet } from 'react-router-dom';

import TopBar from '../components/top_bar/top_bar';

declare global {
    interface Window {
      avatar: string;
      fullName: string;
      email: string;
      logoutUrl: string;
    }
}

const BaseLayout = () => {
  return (
    <>
      <TopBar />
      <Outlet />
    </>
  );
};

export default BaseLayout;
