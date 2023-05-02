// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { BrowserRouter, Route, Routes } from 'react-router-dom';
import NavBar from './features/NavBar';
import ComponentPage from './pages/resources/ComponentPage';

const App = () => {
  return (
    <div className="App">
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<NavBar/>}>
            <Route path="/resources/component/:component" element={<ComponentPage/>}/>
          </Route>
        </Routes>
      </BrowserRouter>
    </div>
  );
};

export default App;
