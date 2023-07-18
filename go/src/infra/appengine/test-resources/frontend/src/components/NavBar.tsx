// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import { Button, Container, Divider, FormControl, Select, SelectChangeEvent } from '@mui/material';
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

  const handleChange = (event: SelectChangeEvent) => {
    componentCtx.api.updateComponent(event.target.value);
  };

  return (
    <Container maxWidth={false}>
      <AppBar>
        <Toolbar>
          <div className={styles.horizontalCenter}>
            <FormControl variant="standard" sx={{ m: 1, minWidth: 120 }}>
              <Select
                data-testid='selectTest'
                value={componentCtx.component}
                onChange={handleChange}
                defaultValue=''
                sx={{ 'color': 'white', '& .MuiSvgIcon-root': {
                  color: 'white',
                }, 'fontSize': '30px', 'minWidth': '250px' }}
                disableUnderline

              >
                {componentCtx.allComponents.length ?
                componentCtx.allComponents.map(
                    (component) =>
                      <MenuItem key={component} value={component}>{component}</MenuItem>,
                ) : null
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
