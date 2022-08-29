// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import './base.css';

import { Link, matchPath, Outlet, useLocation } from 'react-router-dom';

import AppBar from '@mui/material/AppBar';
import Box from '@mui/material/Box';
import Container from '@mui/material/Container';
import Tab from '@mui/material/Tab';
import Tabs from '@mui/material/Tabs';
import Toolbar from '@mui/material/Toolbar';

declare global {
  interface Window {
    avatar: string;
    email: string;
    fullName: string;
    logoutURL: string;
  }
}

function getCurrentLink(linkPatterns: string[]) {
  const { pathname } = useLocation();

  for (let i = 0; i < linkPatterns.length; i += 1) {
    const linkMatch = matchPath(linkPatterns[i], pathname);
    if (linkMatch !== null) {
      return linkMatch;
    }
  }

  return null;
}

export const BaseLayout = () => {
  const linkMatcher = getCurrentLink(['/trigger', '/statistics', '/']);

  var currentTab = '/';
  if (linkMatcher !== null) {
    currentTab = linkMatcher?.pattern?.path;
  }

  return (
    <Box sx={{ flexGrow: 1 }}>
      <AppBar position='static' color='primary'>
        <Toolbar>
          <Tabs value={currentTab} textColor='inherit'>
            <Tab
              className='logoNavTab'
              component={Link}
              label='LUCI Bisection'
              value='/'
              to='/'
            />
            <Tab
              className='navTab'
              component={Link}
              label='New Analysis'
              value='/trigger'
              to='/trigger'
              color='inherit'
              // TODO: remove below once the New Analysis page is implemented
              disabled
            />
            <Tab
              className='navTab'
              component={Link}
              label='Statistics'
              value='/statistics'
              to='/statistics'
              color='inherit'
              // TODO: remove below once the Statistics page is implemented
              disabled
            />
          </Tabs>
          {/* TODO: add login/logout links */}
        </Toolbar>
      </AppBar>
      <Container className='pageContent'>
        <Outlet />
      </Container>
    </Box>
  );
};
