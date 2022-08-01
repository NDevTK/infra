// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import React, { useContext } from 'react';
import { Link } from 'react-router-dom';

import MenuIcon from '@mui/icons-material/Menu';
import Box from '@mui/material/Box';
import IconButton from '@mui/material/IconButton';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';
import Menu from '@mui/material/Menu';
import MenuItem from '@mui/material/MenuItem';
import Typography from '@mui/material/Typography';

import { DynamicComponentNoProps } from '../../../tools/rendering_tools';
import Logo from '../logo/logo';
import { AppBarPage } from '../top_bar';
import { TopBarContext } from '../top_bar_context';

interface Props {
  pages: AppBarPage[];
}

// A menu that is collapsed on mobile.
const CollapsedMenu = ({ pages }: Props) => {
  const { anchorElNav, setAnchorElNav } = useContext(TopBarContext);

  const handleOpenNavMenu = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorElNav(event.currentTarget);
  };

  const handleCloseNavMenu = () => {
    setAnchorElNav(null);
  };

  return (
    <>
      <Box sx={{ flexGrow: 1, display: { xs: 'flex', md: 'none' } }}>
        <IconButton
          size="large"
          aria-label="pages menu"
          aria-controls="menu-appbar"
          aria-haspopup="true"
          onClick={handleOpenNavMenu}
          color="inherit"
          data-testid="collapsed-menu-button"
        >
          <MenuIcon />
        </IconButton>
        <Menu
          id="menu-appbar"
          keepMounted
          anchorEl={anchorElNav}
          anchorOrigin={{
            vertical: 'bottom',
            horizontal: 'left',
          }}
          transformOrigin={{
            vertical: 'top',
            horizontal: 'left',
          }}
          open={Boolean(anchorElNav)}
          onClose={handleCloseNavMenu}
          sx={{
            display: { xs: 'block', md: 'none' },
          }}
          data-testid="collapsed-menu"
        >
          {pages.map((page) => (
            <MenuItem
              component={Link}
              to={page.url}
              key={page.title}
              onClick={handleCloseNavMenu}>
              <ListItemIcon>
                <DynamicComponentNoProps component={page.icon} />
              </ListItemIcon>
              <ListItemText>{page.title}</ListItemText>
            </MenuItem>
          ))}
        </Menu>
      </Box>
      <Box sx={{
        display: {
          xs: 'flex',
          md: 'none' },
        mr: 1,
        width: '3rem',
      }}>
        <Logo />
      </Box>
      <Typography
        variant="h5"
        noWrap
        component="a"
        href=""
        sx={{
          mr: 2,
          display: { xs: 'flex', md: 'none' },
          flexGrow: 1,
          color: 'inherit',
          textDecoration: 'none',
        }}
      >
        Weetbix
      </Typography>
    </>
  );
};

export default CollapsedMenu;
