// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import AppBar from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import { Button, Container, Divider, IconButton, Typography } from '@mui/material';
import { Outlet, useNavigate, useParams } from 'react-router-dom';
import MenuIcon from '@mui/icons-material/Menu';
import TaskAltIcon from '@mui/icons-material/TaskAlt';
import WarningAmberIcon from '@mui/icons-material/WarningAmber';
import ErrorOutlineIcon from '@mui/icons-material/ErrorOutline';
import styles from './NavBar.module.css';

function NavBar() {
  const params = useParams();
  const navigate = useNavigate();

  const updateMetrics = (newComponent : any) => {
    navigate('/' + newComponent + '/component/' + params.component);
  };

  return (
    <Container maxWidth={false}>
      <AppBar>
        <Toolbar>
          <div className={styles.horizontalCenter}>
            <IconButton color="inherit">
              <MenuIcon></MenuIcon>
            </IconButton>
            <Typography className={styles.componentTypography} variant='h6'> {params.component} </Typography>
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
