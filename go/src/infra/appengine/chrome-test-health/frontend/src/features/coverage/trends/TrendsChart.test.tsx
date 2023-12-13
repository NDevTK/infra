import { screen } from '@testing-library/react';
import TrendsChart from './TrendsChart';
import { renderWithContext } from './testUtils';

// Material UI charts need to be mocked since they error
// out due to an export which jest doesn't like.
// Take a look at this github issue:
// https://github.com/mui/material-ui/issues/35465
jest.mock('@mui/x-charts/LineChart', () => (
  { LineChart: jest.fn().mockImplementation(({ children }) => children) }
));

describe('When rendering the trends chart', () => {
  it('Should render loading when isLoading is true and config already loaded', () => {
    renderWithContext(<TrendsChart/>, { data: [], isLoading: true, isConfigLoaded: true });
    expect(screen.getByText('Loading...')).toBeInTheDocument();
  });

  it('Should render the trends chart when data length > 0', () => {
    renderWithContext(<TrendsChart/>, { data: [{ covered: 1, total: 2, date: '' }], isLoading: false, isConfigLoaded: true });
    expect(screen.getByTestId('trendsChart')).toBeInTheDocument();
  });
});
