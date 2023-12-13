import { Box } from '@mui/material';
import { TrendsContextProvider } from '../../features/coverage/trends/TrendsContext';
import TrendsToolbar from '../..//features/coverage/trends/TrendsToolbar';
import TrendsChart from '../../features/coverage/trends/TrendsChart';
import TrendsSearchParams, { PATHS, PLATFORM, PRESETS, UNIT_TESTS_ONLY } from '../../features/coverage/trends/TrendsSearchParams';

function IncrementalTrendsPage() {
  const params = new URLSearchParams(window.location.search);
  const props = {
    platform: params.get(PLATFORM) || '',
    unitTestsOnly: params.get(UNIT_TESTS_ONLY) === 'true',
    paths: params.getAll(PATHS) || [],
    presets: params.getAll(PRESETS) || [],
    isAbsTrend: false,
  };

  return (
    <TrendsContextProvider {...props }>
      <TrendsToolbar />
      <Box sx={{ margin: '10px 20px' }}>
        <TrendsChart />
      </Box>
      <TrendsSearchParams />
    </TrendsContextProvider>
  );
}

export default IncrementalTrendsPage;
