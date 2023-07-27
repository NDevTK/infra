// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import { Button, Checkbox, Container,
  Divider,
  FormControl,
  ListItemText,
  Select,
  SelectChangeEvent,
} from '@mui/material';
import { Outlet, useNavigate, useParams } from 'react-router-dom';
import TaskAltIcon from '@mui/icons-material/TaskAlt';
import WarningAmberIcon from '@mui/icons-material/WarningAmber';
import ErrorOutlineIcon from '@mui/icons-material/ErrorOutline';
import MenuItem from '@mui/material/MenuItem';
import { useContext } from 'react';
import { ComponentContext } from '../features/components/ComponentContext';
import styles from './NavBar.module.css';

function NavBar() {
  const params = useParams();
  const navigate = useNavigate();
  const componentCtx = useContext(ComponentContext);
  const updateMetrics = (newComponent : any) => {
    navigate('/' + newComponent + '/component/' + params.component);
  };

  const handleChange = (event: SelectChangeEvent<typeof componentCtx.allComponents>) => {
    const value = event.target.value;
    componentCtx.api.updateComponents(typeof value === 'string' ? value.split(',') : value);
  };

  return (
    <Container maxWidth={false}>
      <AppBar>
        <Toolbar>
          <div className={styles.horizontalCenter}>
            <FormControl sx={{ 'border': 'none', '& fieldset': { border: 'none' } }}>
              <Select
                data-testid="selectTest"
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
          </div>
          <Divider orientation="vertical" flexItem />
          <div className={styles.horizontalCenter}>
            <Button
              sx={{ my: 2, color: 'white' }}
            >
              <TaskAltIcon className={styles.coverageButton}/>COVERAGE
            </Button>
            <Divider orientation="vertical" flexItem />
            <Button
              sx={{ my: 2, color: 'white' }}
              onClick={() => {
                updateMetrics('resources');
              }}
            >
              <WarningAmberIcon className={styles.resourcesButton}/>RESOURCES
            </Button>
            <Divider orientation="vertical" flexItem />
            <Button
              sx={{ my: 2, color: 'white' }}
            >
              <ErrorOutlineIcon className={styles.flakinessButton}/>FLAKINESS
            </Button>
          </div>
          <Divider orientation="vertical" flexItem />
        </Toolbar>
        <Outlet/>
      </AppBar>
    </Container>

  );
}
export default NavBar;
