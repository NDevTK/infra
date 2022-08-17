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
import Typography from '@mui/material/Typography';

import { AnalysisOverview } from '../../components/analysis_overview/analysis_overview';
import { RevertCLOverview } from '../../components/revert_cl_overview/revert_cl_overview';
import { HeuristicAnalysisTable } from '../../components/heuristic_analysis_table/heuristic_analysis_table';
import { SuspectsOverview } from '../../components/suspects_overview/suspects_overview';
import { getAnalysisDetails } from '../../services/analysis_details';

interface TabPanelProps {
  children?: React.ReactNode;
  name: string;
  value: string;
}

function TabPanel(props: TabPanelProps) {
  const { children, value, name } = props;

  return (
    <div hidden={value !== name} className='tabPanel'>
      {value === name && <div className='tabPanelContents'>{children}</div>}
    </div>
  );
}

export const AnalysisDetailsPage = () => {
  enum AnalysisComponentTabs {
    HEURISTIC = 'Heuristic analysis',
    NTH_SECTION = 'Nth section analysis',
    CULPRIT_VERIFICATION = 'Culprit verification',
  }

  const [currentTab, setCurrentTab] = useState(AnalysisComponentTabs.HEURISTIC);

  const handleTabChange = (
    _: React.SyntheticEvent,
    newTab: AnalysisComponentTabs
  ) => {
    setCurrentTab(newTab);
  };

  const { buildID } = useParams();

  const {
    isLoading,
    isError,
    data: analysisDetails,
  } = useQuery(['analysis', buildID], () => getAnalysisDetails(buildID!));

  if (isLoading) {
    // TODO: update layout so this loading bar spans the entire screen
    return <LinearProgress />;
  }

  if (isError || !analysisDetails) {
    return (
      <main>
        <div className='section'>
          <Alert severity='error'>
            <AlertTitle>Failed to load analysis details</AlertTitle>
            {/* TODO: display more error detail for input issues e.g.
                Build not found, No analysis for that build, etc */}
            An error occurred when querying for the analysis details using build
            ID "{`${buildID}`}".
          </Alert>
        </div>
      </main>
    );
  }

  // TODO: display alert if the build ID queried is not the first failed build
  //       linked to the failure analysis

  return (
    <main>
      <div className='section'>
        <Typography variant='h4' gutterBottom>
          Analysis Details
        </Typography>
        <AnalysisOverview analysis={analysisDetails} />
      </div>
      {analysisDetails.revertCL! && (
        <div className='section'>
          <Typography variant='h4' gutterBottom>
            Revert CL
          </Typography>
          <RevertCLOverview revertCL={analysisDetails.revertCL} />
        </div>
      )}
      {analysisDetails.primeSuspects.length > 0 && (
        <div className='section'>
          <Typography variant='h4' gutterBottom>
            Suspect Summary
          </Typography>
          <SuspectsOverview suspects={analysisDetails.primeSuspects} />
        </div>
      )}
      <div className='section'>
        <Typography variant='h4' gutterBottom>
          Analysis Components
        </Typography>
        <Tabs
          value={currentTab}
          onChange={handleTabChange}
          aria-label='Analysis components tabs'
          className='roundedTabs'
        >
          <Tab
            className='roundedTab'
            value={AnalysisComponentTabs.HEURISTIC}
            label={AnalysisComponentTabs.HEURISTIC}
          />
          <Tab
            className='roundedTab'
            disabled
            value={AnalysisComponentTabs.NTH_SECTION}
            label={AnalysisComponentTabs.NTH_SECTION}
          />
          <Tab
            className='roundedTab'
            disabled
            value={AnalysisComponentTabs.CULPRIT_VERIFICATION}
            label={AnalysisComponentTabs.CULPRIT_VERIFICATION}
          />
        </Tabs>
        <TabPanel value={currentTab} name={AnalysisComponentTabs.HEURISTIC}>
          {/* TODO: Show alert if there are no heuristic results yet */}
          <HeuristicAnalysisTable results={analysisDetails.heuristicResults} />
        </TabPanel>
        <TabPanel value={currentTab} name={AnalysisComponentTabs.NTH_SECTION}>
          {/* TODO: Replace with nth section analysis results */}
          Placeholder for nth section analysis details
        </TabPanel>
        <TabPanel
          value={currentTab}
          name={AnalysisComponentTabs.CULPRIT_VERIFICATION}
        >
          {/* TODO: Replace with culprit verification results */}
          Placeholder for culprit verification details
        </TabPanel>
      </div>
    </main>
  );
};
