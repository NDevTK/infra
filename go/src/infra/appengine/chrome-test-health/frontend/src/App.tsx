// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom';
import { Box } from '@mui/material';
import { useContext } from 'react';
import NavBar from './components/navbar/NavBar';
import TestsPage from './pages/resources/TestsPage';
import { ComponentContextProvider, URL_COMPONENT } from './features/components/ComponentContext';
import { AuthContext } from './features/auth/AuthContext';

const App = () => {
  const { auth } = useContext(AuthContext);
  if (auth === undefined) {
    return null;
  }

  const params = new URLSearchParams(window.location.search);

  let components: string[] = [];
  if (params.has(URL_COMPONENT)) {
    components = params.getAll(URL_COMPONENT);
  } else {
    const local = localStorage.getItem(URL_COMPONENT);
    if (local !== undefined && local !== '' && local !== null) {
      components = local.split(',');
    }
  }

  return (
    <div className="App">
      <BrowserRouter>
        <ComponentContextProvider {...{ components }}>
          <NavBar/>
          <Box component="main" sx={{ flexGrow: 1, minWidth: '1200px' }}>
            <Routes>
              <Route path="/" element={<Navigate to='resources/tests'/>} />
              <Route path="/resources/tests" element={<TestsPage/>} />
            </Routes>
          </Box>
        </ComponentContextProvider>
      </BrowserRouter>
    </div>
  );
};

export default App;
