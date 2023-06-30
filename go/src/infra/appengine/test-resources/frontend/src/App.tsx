// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom';
import { Box } from '@mui/material';
import NavBar from './components/NavBar';
import ResourcesPage from './pages/resources/ResourcesPage';

const App = () => {
  return (
    <div className="App">
      <BrowserRouter>
        <NavBar/>
        <Box component="main" sx={{ flexGrow: 1, marginTop: '74px' }}>
          <Routes>
            <Route path="/" element={<Navigate to='resources/component/Blink>CSS'/>} />
            <Route path="/resources/component/:component" element={<ResourcesPage/>} />
          </Routes>
        </Box>
      </BrowserRouter>
    </div>
  );
};

export default App;
