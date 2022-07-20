// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

import './analysis_details.css';

import { useState } from 'react';
import { useParams } from 'react-router-dom';
import { useQuery } from 'react-query';

import Alert from '@mui/material/Alert';
import AlertTitle from '@mui/material/AlertTitle';
import LinearProgress from '@mui/material/LinearProgress';
import Tab from '@mui/material/Tab';
import Tabs from '@mui/material/Tabs';
import { styled } from '@mui/system';

import AnalysisOverview from '../../components/analysis_overview/analysis_overview';
import ChangeListOverview from '../../components/change_list_overview/change_list_overview';
import HeuristicAnalysisTable from '../../components/heuristic_analysis_table/heuristic_analysis_table';
import SuspectsOverview from '../../components/suspects_overview/suspects_overview';
import { getAnalysisDetails } from '../../services/analysis_details';

const PaperTabs = styled(Tabs)({
  '& .MuiTabs-flexContainer': {
    borderBottom: '1px solid lightgrey',
  },
});

const RoundedTab = styled(Tab)({
  textTransform: 'none',
  paddingLeft: '40px',
  paddingRight: '40px',
  borderTopLeftRadius: '5px',
  borderTopRightRadius: '5px',
  borderLeft: '1px solid lightgrey',
  borderRight: '1px solid lightgrey',
  borderTop: '1px solid lightgrey',
  backgroundColor: 'lightgrey',
  opacity: 0.8,
  '& + &': {
    // add spacing between tabs
    marginLeft: '10px',
  },
  '&.Mui-selected': {
    backgroundColor: 'transparent',
    opacity: '1',
  },
});

const PanelDiv = styled('div')({
  borderBottomLeftRadius: '20px',
  borderBottomRightRadius: '20px',
  border: '1px solid lightgrey',
})

interface TabPanelProps {
  children?: React.ReactNode;
  name: string;
  value: string;
}

function TabPanel(props: TabPanelProps) {
  const { children, value, name, ...other } = props;

  return (
    <PanelDiv
      role='tab-panel'
      hidden={value !== name}
      {...other}
      sx={{
        border: '1px solid lightgrey',
        borderTop: 'none',
      }}
    >
      {value === name && (
        <div className='tabPanel'>
          {children}
        </div>
      )}
    </PanelDiv>
  );
}

const AnalysisDetailsPage = () => {
  const [currentTab, setCurrentTab] = useState('heuristic');

  const handleTabChange = (_: React.SyntheticEvent, newTab: string) => {
    setCurrentTab(newTab);
  };

  const { buildId } = useParams();

  const { isLoading, isError, data: analysisDetails } = useQuery(
    ['analysis', buildId],
    () => getAnalysisDetails(buildId!),
  );

  if (isLoading) {
    return <LinearProgress />;
  }

  if (isError || !analysisDetails) {
    return (
      <main>
        <div className='section'>
          <Alert severity='error'>
            <AlertTitle>Failed to load analysis details</AlertTitle>
            An error occurred when querying for the analysis details using build ID "{`${buildId}`}".
          </Alert>
        </div>
      </main>

    );
  }

  return (
    <main>
      <div className='section'>
        <h1>Analysis Details</h1>
        <AnalysisOverview analysis={analysisDetails} />
      </div>
      <div className='section'>
        {/* TODO: Hide this section if there is no revert CL */}
        <h1>Revert CL</h1>
        <ChangeListOverview changeList={analysisDetails.revertChangeList} />
      </div>
      <div className='section'>
        {/* TODO: Hide this section if there are no suspects */}
        <h1>Suspect Summary</h1>
        <SuspectsOverview suspects={analysisDetails.suspects} />
      </div>
      <div className='section'>
        <h1>Analysis Components</h1>
        <PaperTabs
          value={currentTab}
          onChange={handleTabChange}
          aria-label='Analysis components tabs'
        >
          <RoundedTab
            value='heuristic'
            label='Heuristic analysis'
          />
          <RoundedTab
            value='nth'
            label='Nth section analysis'
          />
          <RoundedTab
            value='culprit'
            label='Culprit verification'
          />
        </PaperTabs>
        <TabPanel
          value={currentTab}
          name='heuristic'
        >
          {/* TODO: Show alert if there are no heuristic results yet */}
          <HeuristicAnalysisTable results={analysisDetails.heuristicAnalysis} />
        </TabPanel>
        <TabPanel
          value={currentTab}
          name='nth'
        >
          {/* TODO: Replace with nth section analysis results */}
          Placeholder for nth section analysis details
        </TabPanel>
        <TabPanel
          value={currentTab}
          name='culprit'
        >
          {/* TODO: Replace with culprit verification results */}
          Placeholder for culprit verification details
        </TabPanel>
      </div>
    </main>
  );
};

export default AnalysisDetailsPage;
