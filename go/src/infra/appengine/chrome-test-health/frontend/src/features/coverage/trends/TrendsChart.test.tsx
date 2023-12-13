import { screen } from '@testing-library/react';
import TrendsChart from './TrendsChart';
import { renderWithContext } from './testUtils';

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
