import * as React from 'react';
import { styled, useTheme, Theme, CSSObject } from '@mui/material/styles';
import Box, { BoxProps as MuiBoxProps } from '@mui/material/Box';
import MuiDrawer from '@mui/material/Drawer';
import MuiAppBar, { AppBarProps as MuiAppBarProps } from '@mui/material/AppBar';
import Toolbar from '@mui/material/Toolbar';
import List from '@mui/material/List';
import CssBaseline from '@mui/material/CssBaseline';
import Typography from '@mui/material/Typography';
import Divider from '@mui/material/Divider';
import IconButton from '@mui/material/IconButton';
import MenuIcon from '@mui/icons-material/Menu';
import ChevronLeftIcon from '@mui/icons-material/ChevronLeft';
import ChevronRightIcon from '@mui/icons-material/ChevronRight';
import ListItemButton from '@mui/material/ListItemButton';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';
import MenuItem from '@mui/material/MenuItem';
import Menu from '@mui/material/Menu';
import Avatar from '@mui/material/Avatar';
import ScienceIcon from '@mui/icons-material/Science';
import HelpIcon from '@mui/icons-material/Help';
import AutoAwesomeMotionIcon from '@mui/icons-material/AutoAwesomeMotion';
import LibraryBooksIcon from '@mui/icons-material/LibraryBooks';
import { Drawer } from '@mui/material';
import { Route, Routes, Link, Navigate } from 'react-router-dom';

import { Asset } from '../asset/Asset';
import { Resource } from '../resource/Resource';

import { AssetList } from '../asset/AssetList';
import { useAppSelector, useAppDispatch } from '../../app/hooks';
import {
  fetchUserPictureAsync,
  logoutAsync,
  setRightSideDrawerClose,
} from './utilitySlice';
import { ResourceList } from '../resource/ResourceList';
import { AssetInstanceList } from '../asset_instance/AssetInstanceList';
import { AssetInstance } from '../asset_instance/AssetInstance';

const drawerWidth = 240;
const rightSideDrawerWidth = 480;

const openedMixin = (theme: Theme): CSSObject => ({
  width: drawerWidth,
  transition: theme.transitions.create('width', {
    easing: theme.transitions.easing.sharp,
    duration: theme.transitions.duration.enteringScreen,
  }),
  overflowX: 'hidden',
});

const closedMixin = (theme: Theme): CSSObject => ({
  transition: theme.transitions.create('width', {
    easing: theme.transitions.easing.sharp,
    duration: theme.transitions.duration.leavingScreen,
  }),
  overflowX: 'hidden',
  width: `calc(${theme.spacing(7)} + 1px)`,
  [theme.breakpoints.up('sm')]: {
    width: `calc(${theme.spacing(8)} + 1px)`,
  },
});

const DrawerHeader = styled('div')(({ theme }) => ({
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'flex-end',
  padding: theme.spacing(0, 1),
  // necessary for content to be below app bar
  ...theme.mixins.toolbar,
}));

interface AppBarProps extends MuiAppBarProps {
  open?: boolean;
}

interface BoxProps extends MuiBoxProps {
  rightSideDrawerOpen?: boolean;
}

const AppBar = styled(MuiAppBar, {
  shouldForwardProp: (prop) => prop !== 'open',
})<AppBarProps>(({ theme, open }) => ({
  zIndex: theme.zIndex.drawer + 1,
  transition: theme.transitions.create(['width', 'margin'], {
    easing: theme.transitions.easing.sharp,
    duration: theme.transitions.duration.leavingScreen,
  }),
  ...(open && {
    marginLeft: drawerWidth,
    width: `calc(100% - ${drawerWidth}px)`,
    transition: theme.transitions.create(['width', 'margin'], {
      easing: theme.transitions.easing.sharp,
      duration: theme.transitions.duration.enteringScreen,
    }),
  }),
}));

const LeftSideDrawer = styled(MuiDrawer, {
  shouldForwardProp: (prop) => prop !== 'open',
})(({ theme, open }) => ({
  width: drawerWidth,
  flexShrink: 0,
  whiteSpace: 'nowrap',
  boxSizing: 'border-box',
  ...(open && {
    ...openedMixin(theme),
    '& .MuiDrawer-paper': openedMixin(theme),
  }),
  ...(!open && {
    ...closedMixin(theme),
    '& .MuiDrawer-paper': closedMixin(theme),
  }),
}));

const CustomBox = styled(Box, {
  shouldForwardProp: (prop) => prop !== 'rightSideDrawerOpen',
})<BoxProps>(({ theme, rightSideDrawerOpen }) => ({
  flexGrow: 1,
  paddingTop: theme.spacing(0),
  paddingLeft: theme.spacing(2),
  paddingRight: theme.spacing(2),
  paddingBottom: theme.spacing(2),
  spacing: theme.spacing(2),
  transition: theme.transitions.create(['width', 'margin'], {
    easing: theme.transitions.easing.sharp,
    duration: theme.transitions.duration.leavingScreen,
  }),
  marginLeft: 0,
  marginRight: 0,
  ...(rightSideDrawerOpen && {
    transition: theme.transitions.create(['width', 'margin'], {
      easing: theme.transitions.easing.easeOut,
      duration: theme.transitions.duration.enteringScreen,
    }),
    marginRight: rightSideDrawerWidth,
  }),
}));

export default function SideDrawerWithAppBar() {
  const theme = useTheme();
  const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);
  const [drawerOpen, setDrawerOpen] = React.useState(false);

  const isMenuOpen = Boolean(anchorEl);
  const menuId = 'primary-search-account-menu';
  const dispatch = useAppDispatch();
  dispatch(fetchUserPictureAsync());
  const userPicture: string = useAppSelector(
    (state) => state.utility.userPicture
  );

  const rightSideDrawerOpen: boolean = useAppSelector(
    (state) => state.utility.rightSideDrawerOpen
  );

  const getActiveEntity: string = useAppSelector(
    (state) => state.utility.activeEntity
  );

  const routes = [
    {
      text: 'Enterprise Lab',
      icon: ScienceIcon,
      path: '/lab',
      entityIdentifier: 'assets',
      component: AssetList,
    },
    {
      text: 'Resources',
      icon: AutoAwesomeMotionIcon,
      path: '/resources',
      entityIdentifier: 'resources',
      component: ResourceList,
    },
    {
      text: 'Lab Instances',
      icon: LibraryBooksIcon,
      path: '/assetInstances',
      entityIdentifier: 'assetInstances',
      component: AssetInstanceList,
    },
  ];

  const handleDrawerOpen = () => {
    setDrawerOpen(true);
  };

  const handleDrawerClose = () => {
    setDrawerOpen(false);
  };

  const handleProfileMenuOpen = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleMenuClose = () => {
    setAnchorEl(null);
  };

  const handleLogout = () => {
    dispatch(logoutAsync());
  };

  const handleRightSideDrawerClose = () => {
    dispatch(setRightSideDrawerClose());
  };

  const renderMenu = (
    <Menu
      anchorEl={anchorEl}
      anchorOrigin={{
        vertical: 'top',
        horizontal: 'right',
      }}
      id={menuId}
      keepMounted
      transformOrigin={{
        vertical: 'top',
        horizontal: 'right',
      }}
      open={isMenuOpen}
      onClose={handleMenuClose}
    >
      <MenuItem onClick={handleLogout}>Logout</MenuItem>
    </Menu>
  );

  const renderRightSideDrawerContents = (activeEntity: string) => {
    switch (activeEntity) {
      case 'assets':
        return <Asset />;
      case 'resources':
        return <Resource />;
      case 'assetInstances':
        return <AssetInstance />;
    }
  };

  return (
    <Box bgcolor={'#dfe3e8'} sx={{ display: 'flex' }}>
      <CssBaseline />
      <AppBar position="fixed" open={drawerOpen}>
        <Toolbar>
          <IconButton
            color="inherit"
            aria-label="open drawer"
            onClick={handleDrawerOpen}
            edge="start"
            sx={{
              marginRight: 5,
              ...(drawerOpen && { display: 'none' }),
            }}
          >
            <MenuIcon />
          </IconButton>
          <Typography variant="h6" noWrap component="div">
            Poros Deployment Manager
          </Typography>
          <Box sx={{ flexGrow: 1 }} />
          <Box sx={{ display: { xs: 'none', md: 'flex' } }}>
            <IconButton
              edge="end"
              aria-label="account of current user"
              aria-controls={menuId}
              aria-haspopup="true"
              onClick={handleProfileMenuOpen}
              color="inherit"
            >
              <Avatar src={userPicture}></Avatar>
            </IconButton>
          </Box>
        </Toolbar>
      </AppBar>
      <LeftSideDrawer variant="permanent" open={drawerOpen} anchor="left">
        <DrawerHeader>
          <IconButton onClick={handleDrawerClose}>
            {theme.direction === 'rtl' ? (
              <ChevronRightIcon />
            ) : (
              <ChevronLeftIcon />
            )}
          </IconButton>
        </DrawerHeader>
        <Divider />

        <List>
          {routes.map((route) => (
            <ListItemButton
              key={route.text}
              sx={{
                minHeight: 48,
                justifyContent: drawerOpen ? 'initial' : 'center',
                px: 2.5,
              }}
              to={route.path}
              component={Link}
            >
              <ListItemIcon
                sx={{
                  minWidth: 0,
                  mr: drawerOpen ? 3 : 'auto',
                  justifyContent: 'center',
                }}
              >
                <route.icon />
              </ListItemIcon>
              <ListItemText
                primary={route.text}
                sx={{ opacity: drawerOpen ? 1 : 0 }}
              />
            </ListItemButton>
          ))}
        </List>
        <Divider />
        <List>
          <ListItemButton
            key="help"
            sx={{
              minHeight: 48,
              justifyContent: drawerOpen ? 'initial' : 'center',
              px: 2.5,
            }}
            component="a"
            href="https://g3doc.corp.google.com/googleclient/chrome/enterprise/g3doc/celab/Poros/user_manual.md"
            rel="noopener"
            target="_blank"
          >
            <ListItemIcon
              sx={{
                minWidth: 0,
                mr: drawerOpen ? 3 : 'auto',
                justifyContent: 'center',
              }}
            >
              <HelpIcon />
            </ListItemIcon>
            <ListItemText primary="Help" sx={{ opacity: drawerOpen ? 1 : 0 }} />
          </ListItemButton>
        </List>
      </LeftSideDrawer>
      {renderMenu}
      <CustomBox rightSideDrawerOpen={rightSideDrawerOpen} component="main">
        <DrawerHeader />
        <Drawer
          sx={{ width: rightSideDrawerWidth }}
          variant="persistent"
          anchor="right"
          open={rightSideDrawerOpen}
        >
          <DrawerHeader />
          <DrawerHeader>
            <IconButton onClick={handleRightSideDrawerClose}>
              <ChevronRightIcon />
            </IconButton>
          </DrawerHeader>
          <Divider />
          {renderRightSideDrawerContents(getActiveEntity)}
        </Drawer>
        <Routes>
          <Route path="/" element={<Navigate to="/lab" />}></Route>
          {routes.map((route) => (
            <Route
              key={route.text}
              element={<route.component />}
              path={route.path}
            ></Route>
          ))}
        </Routes>
      </CustomBox>
    </Box>
  );
}
