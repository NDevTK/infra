// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import AppBar from '@mui/material/AppBar';
import Box from '@mui/material/Box';
import Container from '@mui/material/Container';
import Tab from '@mui/material/Tab';
import Tabs from '@mui/material/Tabs';
import Toolbar from '@mui/material/Toolbar';

import {
  Link,
  matchPath,
  Outlet,
  useLocation
} from 'react-router-dom';

function getCurrentLink(linkPatterns: string[]) {
  const { pathname } = useLocation();

  for (let i = 0; i < linkPatterns.length; i+= 1) {
    const linkMatch = matchPath(linkPatterns[i], pathname)
    if (linkMatch !== null) {
      return linkMatch;
    }
  }

  return null;
}

const BaseLayout = () => {
  const linkMatcher = getCurrentLink([
    '/trigger',
    '/statistics',
    '/'
  ]);

  var currentTab = '/';
  if (linkMatcher !== null) {
    currentTab = linkMatcher?.pattern?.path;
  }

  return (
    <Box sx={{ flexGrow: 1 }}>
      <AppBar position='static' color='primary'>
        <Toolbar>
          <Tabs
            value={currentTab}
            textColor='inherit'
          >
            <Tab
              component={Link}
              label='GoFindit'
              value='/'
              to='/' />
            <Tab
              component={Link}
              label='New Analysis'
              value='/trigger'
              to='/trigger'
              color='inherit' />
            <Tab
              component={Link}
              label='Statistics'
              value='/statistics'
              to='/statistics'
              color='inherit' />
          </Tabs>
          {/* TODO: add login/logout links */}
        </Toolbar>
      </AppBar>
    <Container>
      <Outlet />
    </Container>
    </Box>
  );
};

export default BaseLayout;
