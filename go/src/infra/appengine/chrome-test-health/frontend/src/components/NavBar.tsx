// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import { Checkbox, Container,
  Divider,
  FormControl,
  ListItemText,
  Select,
} from '@mui/material';
import { Outlet } from 'react-router-dom';
import MenuItem from '@mui/material/MenuItem';
import { useContext } from 'react';
import { ComponentContext } from '../features/components/ComponentContext';

function NavBar() {
  const componentCtx = useContext(ComponentContext);

  const handleChange = (event) => {
    const value = event.target.value;
    if (value.length > 0) {
      componentCtx.api.updateComponents(value);
    }
  };

  return (
    <Container maxWidth={false}>
      <AppBar>
        <Toolbar>
          <FormControl sx={{ 'border': 'none', '& fieldset': { border: 'none' } }}>
            <Select
              data-testid="selectComponents"
              multiple
              value={componentCtx.components}
              onChange={handleChange}
              renderValue={(selected) => selected.join(', ')}
              sx={{ 'color': 'white', '& .MuiSvgIcon-root': {
                color: 'white',
              }, 'fontSize': '20px', 'minWidth': '250px', 'maxWidth': '250px' }}
            >
              {componentCtx.allComponents.length ?
                componentCtx.allComponents.map((component) => (
                  <MenuItem key={component} value={component}>
                    <Checkbox checked={componentCtx.components.indexOf(component) > -1} />
                    <ListItemText primary={component} />
                  </MenuItem>
                )) : null
              }
            </Select>
          </FormControl>
          <Divider orientation="vertical" flexItem />
        </Toolbar>
        <Outlet/>
      </AppBar>
    </Container>

  );
}
export default NavBar;
