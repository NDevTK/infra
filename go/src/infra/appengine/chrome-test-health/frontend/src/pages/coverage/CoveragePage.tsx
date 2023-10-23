import { Box } from '@mui/material';
import { SummaryContextProvider } from '../../features/coverage/summary/SummaryContext';
import SummaryTable from '../../features/coverage/summary/SummaryTable';
import SummaryToolbar from '../../features/coverage/summary/SummaryToolbar';
import SummarySearchParams, { REVISION, UNIT_TESTS_ONLY, PLATFORM } from '../../features/coverage/summary/SummarySearchParams';

function CoveragePage() {
  const params = new URLSearchParams(window.location.search);
  const props = {
    revision: params.get(REVISION) || '',
    platform: params.get(PLATFORM) || '',
    unitTestsOnly: params.get(UNIT_TESTS_ONLY) === 'true',
  };

  return (
    <SummaryContextProvider {...props}>
      <SummaryToolbar/>
      <Box sx={{ margin: '10px 20px' }}>
        <SummaryTable />
      </Box>
      <SummarySearchParams />
    </SummaryContextProvider>
  );
}

export default CoveragePage;
