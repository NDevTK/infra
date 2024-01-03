// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import {
  Autocomplete,
  AutocompleteChangeReason,
  Button,
  Divider,
  FilterOptionsState,
  TextField,
} from '@mui/material';
import { Link, Outlet } from 'react-router-dom';
import { useCallback, useContext, useRef, useState } from 'react';
import { ComponentContext } from '../../features/components/ComponentContext';
import styles from './NavBar.module.css';

function filterComponents(all: string[], state: FilterOptionsState<string>): string[] {
  const pieces = state.inputValue.split(',');
  const fragment = (pieces.pop() || '').trim().toLowerCase();
  return all.filter((cmp) => cmp.toLowerCase().includes(fragment));
}

function sanitizeComponents(current: string, add: string | null = null): string {
  let newComponents = current.split(',');
  if (add !== null && add !== '') {
    // If a new value has been selected from the autocomplete menu, we are most
    // likely replacing the last value.
    // i.e. "Blink, CSS" => "Blink, Blink>CSS"
    newComponents.pop();
    newComponents.push(add);
  }
  newComponents = newComponents.map((c) => c.trim());
  newComponents = newComponents.filter((c, i) => {
    // Remove all empty and duplicate values
    return c !== '' && newComponents.indexOf(c) === i;
  });
  return newComponents.join(', ');
}

function NavBar() {
  const componentCtx = useContext(ComponentContext);
  const [components, setComponents] = useState(componentCtx.components?.join(', ') || '');
  const componentsFieldRef = useRef<HTMLInputElement>(null);

  const onComponentsChange = useCallback((
      _: React.SyntheticEvent<Element, Event>,
      value: string | null,
      reason: AutocompleteChangeReason,
  ) => {
    switch (reason) {
      case 'clear':
        setComponents('');
        break;
      case 'createOption':
        setComponents(sanitizeComponents(value || ''));
        componentsFieldRef.current?.blur();
        break;
      case 'selectOption': {
        const newComponents = sanitizeComponents(
            componentsFieldRef.current?.value || '',
            value,
        );
        setComponents(newComponents);
        break;
      }
    }
  }, [setComponents]);

  const onComponentsBlur = useCallback(() => {
    // We have to use the field ref as components may not have set yet
    const components = sanitizeComponents(componentsFieldRef.current?.value || '');
    setComponents(components);
    if (components === '') {
      componentCtx.api.updateComponents([]);
    } else {
      componentCtx.api.updateComponents(components.split(', '));
    }
  }, [componentCtx, setComponents]);

  return (
    <AppBar position='relative'>
      <Toolbar>
        <Autocomplete
          freeSolo
          data-testid="componentsAutocomplete"
          options={componentCtx.allComponents}
          value={components}
          onChange={onComponentsChange}
          onBlur={onComponentsBlur}
          filterOptions={filterComponents}
          renderInput={(params) => (
            <TextField {...params}
              variant='standard'
              data-testid="componentsTextField"
              className={styles.componentInput}
              placeholder='All Components'
              inputRef={componentsFieldRef}
              InputProps={{
                disableUnderline: true,
                style: {
                  color: 'white',
                  fontSize: '1.25rem',
                  fontWeight: 400,
                },
                ...params.InputProps,
              }}
              sx={{ input: { '&::placeholder': { opacity: 0.9 } } }}
            />
          )}
          sx={{
            'width': '500px',
            'mr': 2,
            '& .MuiSvgIcon-root': {
              color: 'white',
            },
          }}
        />
        <Divider orientation="vertical" flexItem />
        <Button disableElevation component={Link} to="/coverage/summary" variant="contained" color="primary">
          Coverage
        </Button>
        <Divider orientation="vertical" flexItem />
        <Button disableElevation component={Link} to="/resources/tests" variant="contained" color="primary">
          Resources
        </Button>
        <Divider orientation="vertical" flexItem />
      </Toolbar>
      <Outlet/>
    </AppBar>
  );
}
export default NavBar;
