// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import { useContext, useEffect, useState } from 'react';
import { Autocomplete, TextField, Toolbar } from '@mui/material';
import { Team, getTeams } from '../../../api/coverage';
import { AuthContext } from '../../../features/auth/AuthContext';
import { ComponentContext } from '../../../features/components/ComponentContext';

export interface TeamItem {
  label: string,
  team: Team
}

function TeamsToolbar() {
  const { api } = useContext(ComponentContext);
  const { auth } = useContext(AuthContext);
  const [teams, setTeams] = useState([] as Team[]);
  const [autocompleteInputValue, setAutocompleteInputValue] = useState('');

  useEffect(() => {
    if (auth != undefined) {
      getTeams(auth).then((response) => {
        setTeams(response.teams);
      });
    }
  }, []);

  return (
    <>
      <Toolbar sx={{ mt: 1 }}>
        <Autocomplete
          options={teams}
          getOptionLabel={(option: Team) => option.name}
          sx={{ width: 300 }}
          renderInput={(params) => <TextField {...params} label="Select Team" variant="standard" />}
          onChange={(_, selectedTeam: Team | null) => {
            if (selectedTeam) {
              api.updateComponents(selectedTeam.components);
            } else {
              api.updateComponents([]);
            }
          }}
          inputValue={autocompleteInputValue}
          onInputChange={(_, val) => {
            setAutocompleteInputValue(val);
          }}
          data-testid="teamsAutocompleteTest"
        />
      </Toolbar>
    </>
  );
}

export default TeamsToolbar;
