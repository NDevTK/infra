import { ReactElement } from 'react';
import { render } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import * as SummaryContext from '../../features/coverage/summary/SummaryContext';
import CoveragePage from './CoveragePage';

export function renderWithBrowserRouter(
    ui: ReactElement,
) {
  render(
      <BrowserRouter>
        {ui}
      </BrowserRouter>,
  );
}

describe('when rendering the CoveragePage', () => {
  // This is needed to allow us to modify window.location
  Object.defineProperty(window, 'location', {
    writable: true,
    value: { assign: jest.fn() },
  });

  it('should pass in default values', async () => {
    const mockContext = jest.fn();
    jest.spyOn(SummaryContext, 'SummaryContextProvider').mockImplementation((props) => {
      return mockContext(props);
    });
    renderWithBrowserRouter(<CoveragePage/>);
    expect(mockContext).toHaveBeenCalledWith(
        expect.objectContaining({
          revision: '',
          platform: '',
          unitTestsOnly: false,
        }),
    );
  });

  it('should pass in url param values', async () => {
    const mockContext = jest.fn();
    jest.spyOn(SummaryContext, 'SummaryContextProvider').mockImplementation((props) => {
      return mockContext(props);
    });
    window.location.search = 'https://localhost/?placeholder'+
    '=placeholder&rev=abcd124&plat=linux&isTest=true';
    renderWithBrowserRouter(<CoveragePage/>);
    expect(mockContext).toHaveBeenCalledWith(
        expect.objectContaining({
          revision: 'abcd124',
          platform: 'linux',
          unitTestsOnly: true,
        }),
    );
  });
});
