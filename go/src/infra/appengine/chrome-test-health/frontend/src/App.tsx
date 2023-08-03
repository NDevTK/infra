// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom';
import { Box } from '@mui/material';
import NavBar from './components/NavBar';
import ResourcesPage from './pages/resources/ResourcesPage';
import { ComponentContextProvider } from './features/components/ComponentContext';
import { COMPONENT } from './features/components/ComponentParams';

const App = () => {
  const params = new URLSearchParams(window.location.search);
  const components = params.has(COMPONENT) ?
    params.getAll(COMPONENT) :
    localStorage.getItem(COMPONENT)?.split(',') || ['Blink'];
  return (
    <div className="App">
      <BrowserRouter>
        <ComponentContextProvider {...{ components }}>
          <NavBar/>
          <Box component="main" sx={{ flexGrow: 1, marginTop: '74px' }}>
            <Routes>
              <Route path="/" element={<Navigate to='resources/tests'/>} />
              <Route path="/resources/tests" element={<ResourcesPage/>} />
            </Routes>
          </Box>
        </ComponentContextProvider>
      </BrowserRouter>

    </div>
  );
};

export default App;
