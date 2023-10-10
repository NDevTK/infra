import { Box } from '@mui/material';
import { SummaryContextProvider } from '../../features/coverage/summary/SummaryContext';
import SummaryTable from '../../features/coverage/summary/SummaryTable';
import SummaryToolbar from '../../features/coverage/summary/SummaryToolbar';
import TeamsToolbar from '../../features/coverage/summary/TeamsToolbar';

function CoveragePage() {
  const params = new URLSearchParams(window.location.search);
  const props = {
    revision: params.get('revision') || '',
    platform: params.get('platform') || '',
    unitTestsOnly: params.get('unit_tests_only') === 'true',
  };

  return (
    <SummaryContextProvider {...props}>
      <TeamsToolbar />
      <SummaryToolbar/>
      <Box sx={{ margin: '10px 20px' }}>
        <SummaryTable />
      </Box>
    </SummaryContextProvider>
  );
}

export default CoveragePage;
