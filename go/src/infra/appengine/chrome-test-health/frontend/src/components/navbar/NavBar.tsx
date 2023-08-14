// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import {
  Autocomplete,
  Divider,
  TextField,
} from '@mui/material';
import { Outlet } from 'react-router-dom';
import { useContext, useState } from 'react';
import { ComponentContext } from '../../features/components/ComponentContext';

function NavBar() {
  const componentCtx = useContext(ComponentContext);
  const [navComps, setNavComps] = useState(componentCtx.components);

  const handleChange = (_, components) => {
    setNavComps(components.filter(
        (component) => component !== 'All Components'));
  };

  const handleBlur = () => {
    componentCtx.api.updateComponents(navComps);
  };

  function displayComponents(components: string[]) {
    const finalDisplay: JSX.Element[] = [];
    components.forEach((component, index) => {
      finalDisplay.push(
          <p style={{ minWidth: 'fit-content' }} key={component}>
            {index !== 0 ? ', ' : ''}
            {component}
          </p>,
      );
    });
    return finalDisplay;
  }

  return (
    <AppBar position='relative'>
      <Toolbar>
        <Autocomplete
          multiple
          limitTags={3}
          data-testid="selectComponents"
          options={componentCtx.allComponents}
          value={navComps.length > 0 ? navComps : ['All Components']}
          onChange={handleChange}
          onBlur={handleBlur}
          disableCloseOnSelect
          getLimitTagsText={(more) => `... +${more}`}
          renderInput={(params) => (
            <TextField {...params} InputProps={{ ...params.InputProps, style: { color: 'white' } }}/>
          )}
          renderTags={(components) => displayComponents(components)}
          sx={{ 'width': '700px', 'border': 'none', '& fieldset': { border: 'none' },
            '& .MuiSvgIcon-root': {
              color: 'white',
            }, '& .MuiAutocomplete-tag': {
              color: 'white',
            },
            '& .MuiAutocomplete-inputRoot': {
              flexWrap: 'nowrap',
              overflowX: 'auto',
              overflowY: 'hidden',
              maxHeight: '73px',
            },
            '& .MuiAutocomplete-endAdornment': {
              position: 'sticky',
              flex: 'none',
              transform: 'translateX(50px)',
            },
            '& .MuiIconButton-root': {
              backgroundColor: 'rgb(49 49 50)',
              margin: '5px',
              maxHeight: '20px',
              maxWidth: '20px',
            },
          }}
        />
        <Divider orientation="vertical" flexItem />
      </Toolbar>
      <Outlet/>
    </AppBar>
  );
}
export default NavBar;
